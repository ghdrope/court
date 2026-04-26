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

	"github.com/ghdrope/court/internal/court"
	"github.com/ghdrope/court/pkg/redis"
	"go.uber.org/zap"
)

// SuitCloseRequestedEvent represents a request to close a suit.
type SuitCloseRequestedEvent struct {
	// IncidentID identifies the incident associated with the suit.
	IncidentID string `json:"incident_id"`

	// Reason explains why the suit should be closed.
	Reason string `json:"reason"`
}

// SuitCloseConsumer consumes suit.close.requested events and closes suits.
type SuitCloseConsumer struct {
	Client *redis.StreamClient
	Log    *zap.Logger
}

// NewSuitCloseConsumer creates a new SuitCloseConsumer.
func NewSuitCloseConsumer(client *redis.StreamClient, log *zap.Logger) *SuitCloseConsumer {
	return &SuitCloseConsumer{
		Client: client,
		Log:    log,
	}
}

// Start begins consuming suit.close.requested events and processing them.
func (c *SuitCloseConsumer) Start(ctx context.Context, svc *court.Service) error {

	logger := c.Log.With(zap.String("component", "suit-close-consumer"))

	if err := c.Client.EnsureGroup(ctx); err != nil {
		logger.Error("failed to initialize stream consumer group", zap.Error(err))
		return err
	}

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var evt SuitCloseRequestedEvent

		if err := json.Unmarshal(data, &evt); err != nil {
			logger.Error("received invalid event payload", zap.Error(err))
			return err
		}

		if evt.IncidentID == "" {
			logger.Warn("received event without incident ID")
			return nil
		}

		logger.Info("processing suit close request",
			zap.String("incident_id", evt.IncidentID),
			zap.String("reason", evt.Reason),
		)

		if err := svc.CloseTrial(ctx, evt.IncidentID, evt.Reason); err != nil {
			logger.Error("failed to close suit",
				zap.String("incident_id", evt.IncidentID),
				zap.Error(err),
			)
			return err
		}

		logger.Info("suit successfully closed",
			zap.String("incident_id", evt.IncidentID),
		)

		return nil
	})
}
