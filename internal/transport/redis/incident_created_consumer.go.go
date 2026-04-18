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

	"github.com/ghdrope/court/internal/prosecutor"
	"github.com/ghdrope/court/pkg/redis"
	"go.uber.org/zap"
)

// Stream configuration for incident processing.
//
// These events are produced by the Officer service when
// a new IncidentReport is persisted in PostgreSQL.
const (
	ProsecutorStream = "incident.created"
	ProsecutorGroup  = "prosecutor-group"
)

// IncidentCreatedEvent represents the payload of incident.created events.
type IncidentCreatedEvent struct {
	IncidentID string `json:"incident_id"`
}

// IncidentCreatedConsumer handles incident.created events.
type IncidentCreatedConsumer struct {
	Client *redis.StreamClient
	Log    *zap.Logger
}

// ConsumeIncidentCreation starts the event loop for processing incident.created events.
//
// Messages are acknowledged only after successful processing.
func NewIncidentCreatedConsumer(client *redis.StreamClient, log *zap.Logger) *IncidentCreatedConsumer {
	return &IncidentCreatedConsumer{
		Client: client,
		Log:    log,
	}
}

// StartIncidentCreatedConsumer
func (c *IncidentCreatedConsumer) Start(ctx context.Context, svc *prosecutor.Service) error {

	logger := c.Log.With(zap.String("component", "prosecutor-consumer"))

	if err := c.Client.EnsureGroup(ctx); err != nil {
		return err
	}

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var evt IncidentCreatedEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			logger.Error("invalid payload", zap.Error(err))
			return err
		}

		if evt.IncidentID == "" {
			logger.Error("missing incident_id")
			return nil
		}

		logger.Info("event received",
			zap.String("incident_id", evt.IncidentID),
		)

		inc, err := svc.Repo.GetByID(ctx, evt.IncidentID)
		if err != nil {
			logger.Error("failed to fetch incident", zap.Error(err))
			return err
		}

		// Process incident
		logger.Info("starting incident analysis",
			zap.String("incident_id", evt.IncidentID),
		)
		if err := svc.ProcessIncident(ctx, inc); err != nil {
			logger.Error("failed to process incident", zap.Error(err))
			return err
		}

		logger.Info("incident processed",
			zap.String("incident_id", evt.IncidentID),
		)

		return nil
	})
}
