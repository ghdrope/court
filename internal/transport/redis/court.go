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
	"github.com/ghdrope/court/internal/court"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Court stream and consumer configuration.
const (
	CourtStream   = "court:stream"
	CourtGroup    = "court-group"
	CourtConsumer = "court-1"
)

// CourtStreamClient publishes prosecutor completion events.
type CourtStreamClient struct {
	*Client
}

// NewCourtStreamClient creates a new court stream publisher.
func NewCourtStreamClient(base *Client) *CourtStreamClient {
	return &CourtStreamClient{Client: base}
}

// PublishedStored emits a prosecutor.finished event.
func (c *CourtStreamClient) PublishStored(
	ctx context.Context,
	event archive.StoredEvent,
) error {

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal court event: %w", err)
	}

	if err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: CourtStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err(); err != nil {
		return fmt.Errorf("xadd court event: %w", err)
	}

	return nil
}

// EnsureCourtGroup ensures Redis consumer group exists.
func (c *Client) EnsureCourtGroup(ctx context.Context) error {

	err := c.rdb.XGroupCreateMkStream(
		ctx,
		CourtStream,
		CourtGroup,
		"0",
	).Err()

	if err != nil && err != redis.Nil && !isBusyGroup(err) {
		return fmt.Errorf("create court group: %w", err)
	}

	return nil
}

// ConsumerCourtloop processes prosecutor.finished events
// and creates formal legal cases.
func (c *Client) ConsumeCourtLoop(
	ctx context.Context,
	svc *court.Service,
) error {

	logger := zap.L().With(zap.String("component", "court-consumer"))

	for {
		res, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    CourtGroup,
			Consumer: CourtConsumer,
			Streams:  []string{CourtStream, ">"},
			Count:    10,
			Block:    0,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			logger.Error("failed reading court stream", zap.Error(err))
			continue
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {

				raw, ok := msg.Values["payload"].(string)
				if !ok {
					logger.Warn("invalid payload type")
					continue
				}

				var event archive.StoredEvent
				if err := json.Unmarshal([]byte(raw), &event); err != nil {
					logger.Error("invalid event payload", zap.Error(err))
					continue
				}

				// Only handle prosecutor completion events
				if event.Type != "prosecutor.finished" {
					continue
				}

				payloadBytes, err := json.Marshal(event.Payload)
				if err != nil {
					logger.Error("marshal payload", zap.Error(err))
					continue
				}

				var data struct {
					IncidentID string `json:"incident_id"`
					Cluster    string `json:"cluster"`
					Namespace  string `json:"namespace"`
					Pod        string `json:"pod"`
				}

				if err := json.Unmarshal(payloadBytes, &data); err != nil {
					logger.Error("invalid court payload", zap.Error(err))
					continue
				}

				if err := svc.CreateSuit(
					ctx,
					data.IncidentID,
					data.Cluster,
					data.Namespace,
					data.Pod,
				); err != nil {
					logger.Error("failed to create case", zap.Error(err))
					continue
				}

				if err := c.rdb.XAck(ctx, CourtStream, CourtGroup, msg.ID).Err(); err != nil {
					logger.Error("failed to ACK", zap.Error(err))
				}
			}
		}
	}
}
