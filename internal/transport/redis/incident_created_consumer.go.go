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
	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/pkg/redis"
	"go.uber.org/zap"
)

// Stream configuration for incident processing.
//
// These events are produced by the Officer service when
// a new IncidentReport is persisted in PostgreSQL.
const (
	IncidentCreatedStream = "incident.created"
	CourtGroup            = "court-group"
)

// IncidentCreatedEvent is emitted by Officer when a new incident is created.
type IncidentCreatedEvent struct {
	IncidentID string `json:"incident_id"`
}

// IncidentRepository defines access to incident data.
type IncidentRepository interface {
	GetByID(ctx context.Context, id string) (*incident.IncidentReport, error)
}

// IncidentCreatedConsumer handles incident.created events.
type IncidentCreatedConsumer struct {
	Client *redis.StreamClient
	Repo   IncidentRepository
	Log    *zap.Logger
}

// NewIncidentCreatedConsumer creates consumer.
func NewIncidentCreatedConsumer(client *redis.StreamClient, repo IncidentRepository, log *zap.Logger) *IncidentCreatedConsumer {
	return &IncidentCreatedConsumer{
		Client: client,
		Repo:   repo,
		Log:    log,
	}
}

// StartIncidentCreatedConsumer begins consuming incident.created events.
func (c *IncidentCreatedConsumer) Start(ctx context.Context, svc *court.Service) error {

	logger := c.Log.With(zap.String("component", "court-consumer"))

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

		logger.Info("incident.created received",
			zap.String("incident_id", evt.IncidentID),
		)

		incident, err := c.Repo.GetByID(ctx, evt.IncidentID)
		if err != nil {
			logger.Error("failed to fetch incident", zap.Error(err))
			return err
		}

		return svc.CreateSuit(ctx, incident)
	})
}
