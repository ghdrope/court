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
	"github.com/ghdrope/court/internal/prosecutor"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Stream and consumer configuration for the prosecutor pipeline.
const (
	ProsecutorStream   = "prosecutor:stream"
	ProsecutorGroup    = "prosecutor-group"
	ProsecutorConsumer = "prosecutor-1"
)

// ProsecutorStreamClient publishes stored events
// to Prosecutor service.
type ProsecutorStreamClient struct {
	rdb *goredis.Client
}

// NewProsecutorStreamClient creates a new instance.
func NewProsecutorStreamClient(rdb *goredis.Client) *ProsecutorStreamClient {
	return &ProsecutorStreamClient{rdb: rdb}
}

// PublishStored publishes a generic stored event.
func (c *ProsecutorStreamClient) PublishStored(
	ctx context.Context,
	event archive.StoredEvent,
) error {

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal stored event: %w", err)
	}

	if err := c.rdb.XAdd(ctx, &goredis.XAddArgs{
		Stream: ProsecutorStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err(); err != nil {
		return fmt.Errorf("xadd prosecutor event: %w", err)
	}

	return nil
}

// EnsureProsecutorGroup ensures that the Redis consumer group exists
// for the prosecutor stream.
func EnsureProsecutorGroup(ctx context.Context, rdb *goredis.Client) error {

	err := rdb.XGroupCreateMkStream(
		ctx,
		ProsecutorStream,
		ProsecutorGroup,
		"0",
	).Err()

	if err != nil && err != goredis.Nil && !isBusyGroup(err) {
		return fmt.Errorf("create prosecutor group: %w", err)
	}

	return nil
}

// ConsumeProsecutorLoop consumes consumes StoredEvent messages
// from Redis Stream and delegates processing to the Prosecutor.
func (c *ProsecutorStreamClient) ConsumeProsecutorLoop(
	ctx context.Context,
	svc *prosecutor.Service,
) error {

	logger := zap.L().With(zap.String("component", "prosecutor-consumer"))

	for {
		res, err := c.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
			Group:    ProsecutorGroup,
			Consumer: ProsecutorConsumer,
			Streams:  []string{ProsecutorStream, ">"},
			Count:    10,
			Block:    0,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			logger.Error("failed reading stream", zap.Error(err))
			continue
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {

				raw, ok := msg.Values["payload"].(string)
				if !ok {
					logger.Warn("invalid payload")
					continue
				}

				var event archive.StoredEvent
				if err := json.Unmarshal([]byte(raw), &event); err != nil {
					logger.Error("invalid event", zap.Error(err))
					continue
				}

				// Only handle incident events
				if event.Type != "incident.stored" {
					continue
				}

				payloadBytes, err := json.Marshal(event.Payload)
				if err != nil {
					logger.Error("marshal payload", zap.Error(err))
				}

				var r incident.IncidentReport
				if err := json.Unmarshal(payloadBytes, &r); err != nil {
					logger.Error("invalid incident payload", zap.Error(err))
					continue
				}

				// Process incident
				if err := svc.ProcessIncident(ctx, &r); err != nil {
					logger.Error("failed processing incident", zap.Error(err))
				}

				// ACK only after success
				if err := c.rdb.XAck(ctx, ProsecutorStream, msg.ID).Err(); err != nil {
					logger.Error("failed to ACK", zap.Error(err))
				}
			}
		}
	}
}
