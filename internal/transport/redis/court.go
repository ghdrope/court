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
	"github.com/ghdrope/court/internal/court"
	"github.com/ghdrope/court/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Stream configuration for court processing.
const (
	CourtStream   = "court:stream"
	CourtGroup    = "court-group"
	CourtConsumer = "court-1"
)

// CourtStreamClient handles publishing and consuming court-related events.
type CourtStreamClient struct {
	Client *redis.StreamClient
}

// NewCourtStreamClient creates a new CourtStreamClient.
func NewCourtStreamClient(rdb *goredis.Client) *CourtStreamClient {
	return &CourtStreamClient{
		Client: redis.NewStreamClient(rdb, redis.Config{
			Stream:   CourtStream,
			Group:    CourtGroup,
			Consumer: CourtConsumer,
		}),
	}
}

// PublishedStored publishes a prosecutor completion event into the stream.
func (c *CourtStreamClient) PublishStored(ctx context.Context, event archive.StoredEvent) error {

	return c.Client.Publish(ctx, event)
}

// ConsumerCourt processes prosecutor.finished events and creates court cases.
//
// Only events of type "prosecutor.finished" are handled.
// Successful processing results in message acknowledgement.
func (c *CourtStreamClient) ConsumeCourt(ctx context.Context, svc *court.Service) error {

	logger := zap.L().With(zap.String("component", "court-consumer"))

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var event archive.StoredEvent
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("invalid event payload", zap.Error(err))
			return err
		}

		// Filter only relevant events
		if event.Type != "prosecutor.finished" {
			return nil
		}

		payloadBytes, err := json.Marshal(event.Payload)
		if err != nil {
			logger.Error("failed to marshal payload", zap.Error(err))
			return err
		}

		var payload struct {
			IncidentID string `json:"incident_id"`
			Cluster    string `json:"cluster"`
			Namespace  string `json:"namespace"`
			Pod        string `json:"pod"`
		}

		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			logger.Error("invalid court payload", zap.Error(err))
			return err
		}

		if err := svc.CreateSuit(
			ctx,
			payload.IncidentID,
			payload.Cluster,
			payload.Namespace,
			payload.Pod,
		); err != nil {
			logger.Error("failed to create suit", zap.Error(err))
			return err
		}

		return nil
	})
}
