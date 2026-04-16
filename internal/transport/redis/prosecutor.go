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

	"github.com/ghdrope/court/internal/archive"
	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/prosecutor"
	"github.com/ghdrope/court/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Stream configuration for prosecutor processing.
const (
	ProsecutorStream   = "prosecutor:stream"
	ProsecutorGroup    = "prosecutor-group"
	ProsecutorConsumer = "prosecutor-1"
)

// ProsecutorStreamClient handles publishing and consuming prosecutor events.
type ProsecutorStreamClient struct {
	Client *redis.StreamClient
}

// NewProsecutorStreamClient creates a new ProsecutorStreamClient.
func NewProsecutorStreamClient(rdb *goredis.Client) *ProsecutorStreamClient {
	return &ProsecutorStreamClient{
		Client: redis.NewStreamClient(rdb, redis.Config{
			Stream:   ProsecutorStream,
			Group:    ProsecutorGroup,
			Consumer: ProsecutorConsumer,
		}),
	}
}

// PublishStored publishes a stored event into the stream.
func (c *ProsecutorStreamClient) PublishStored(ctx context.Context, event archive.StoredEvent) error {
	return c.Client.Publish(ctx, event)
}

// ConsumeProsecutor processes stored events and triggers prosecutor logic.
//
// Only events of type "incident.stored" are handled.
// Messages are acknowledged only after successful processing.
func (c *ProsecutorStreamClient) ConsumeProsecutor(ctx context.Context, svc *prosecutor.Service) error {

	logger := zap.L().With(zap.String("component", "prosecutor-consumer"))

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var event archive.StoredEvent
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("invalid event payload", zap.Error(err))
			return err
		}

		// Filter only relevant events
		if event.Type != "incident.stored" {
			return nil
		}

		payloadBytes, err := json.Marshal(event.Payload)
		if err != nil {
			logger.Error("failed to marshal payload", zap.Error(err))
			return err
		}

		var r incident.IncidentReport
		if err := json.Unmarshal(payloadBytes, &r); err != nil {
			logger.Error("invalid incident payload", zap.Error(err))
			return err
		}

		// Process incident
		if err := svc.ProcessIncident(ctx, &r); err != nil {
			logger.Error("failed incident payload", zap.Error(err))
			return err
		}

		return nil
	})
}
