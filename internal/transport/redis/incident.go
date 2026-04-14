/*
Copyright 2026 Pedro Cozinheiro.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package redisstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ghdrope/court/internal/archive"
	"github.com/ghdrope/court/internal/incident"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Stream and consumer group configuration for incidents.
const (
	IncidentStream   = "incident:stream"
	IncidentGroup    = "incident-archive-group"
	IncidentConsumer = "archive-1"
)

// IncidentStreamClient implements publishing and consuming logic
// for IncidentReport events.
type IncidentStreamClient struct {
	*Client
}

// NewIncidentStreamClient creates a new incident stream client.
func NewIncidentStreamClient(base *Client) *IncidentStreamClient {
	return &IncidentStreamClient{Client: base}
}

// PublishIncident publishes an IncidentReport into Redis Stream.
func (c *IncidentStreamClient) PublishIncident(
	ctx context.Context,
	r *incident.IncidentReport,
) error {

	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal incident: %w", err)
	}

	if err = c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: IncidentStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err(); err != nil {
		return fmt.Errorf("xadd incident: %w", err)
	}

	return nil
}

// Send implements officer.ArchiveClient interface.
// Keeps Officer decoupled from transport implementation.
func (c *IncidentStreamClient) Send(
	ctx context.Context,
	r *incident.IncidentReport,
) error {
	return c.PublishIncident(ctx, r)
}

// EnsureGroup ensures that the Redis consumer group exists.
// Safe to call multiple times.
func (c *IncidentStreamClient) EnsureGroup(ctx context.Context) error {

	err := c.rdb.XGroupCreateMkStream(
		ctx,
		IncidentStream,
		IncidentGroup,
		"0",
	).Err()

	if err != nil && err != redis.Nil && !isBusyGroup(err) {
		return fmt.Errorf("create group: %w", err)
	}

	return nil
}

// isBusyGroup checks if group already exists error.
func isBusyGroup(err error) bool {
	return err != nil && err.Error() == "BUSYGROUP Consumer Group name already exists"
}

// ConsumeLoop continuously consumes incident events from Redis Stream
// and persists them into the archive database.
func (c *IncidentStreamClient) ConsumeLoop(
	ctx context.Context,
	arch *archive.Archive,
) error {

	logger := zap.L().With(zap.String("component", "incident-stream"))

	for {
		res, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    IncidentGroup,
			Consumer: IncidentConsumer,
			Streams:  []string{IncidentStream, ">"},
			Count:    10,
			Block:    0,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			logger.Error("failed reading Redis stream", zap.Error(err))
			continue
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {

				raw, ok := msg.Values["payload"].(string)
				if !ok {
					logger.Warn("invalid payload type", zap.String("msg_id", msg.ID))
					continue
				}

				var r incident.IncidentReport
				if err := json.Unmarshal([]byte(raw), &r); err != nil {
					logger.Error("invalid incident payload", zap.Error(err))
					continue
				}

				// Persist incident into PostgreSQL archive.
				if err := arch.StoreIncident(ctx, &r); err != nil {
					logger.Error("failed to persist incident", zap.Error(err))
				}

				// Acknowledge only after successful persistence.
				if err := c.rdb.XAck(ctx, IncidentStream, IncidentGroup, msg.ID).Err(); err != nil {
					logger.Error("failed to ACK message", zap.Error(err))
				}

			}
		}
	}
}
