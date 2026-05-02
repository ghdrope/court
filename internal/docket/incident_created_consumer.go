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

package docket

import (
	"context"
	"encoding/json"

	"github.com/ghdrope/court/internal/court"
	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/pkg/redis"
	"go.uber.org/zap"
)

// IncidentCreatedEvent represents the payload emitted when a new incident is created.
type IncidentCreatedEvent struct {
	// IncidentID is the unique identifier of the incident.
	IncidentID string `json:"incident_id"`
}

// IncidentRepository defines access to incident data.
type IncidentRepository interface {
	// GetByID retrieves an incident by its ID.
	GetByID(ctx context.Context, id string) (*incident.IncidentReport, error)
}

// IncidentCreatedConsumer consumes incident.created events and triggers suit creation.
type IncidentCreatedConsumer struct {
	Client *redis.StreamClient
	Repo   IncidentRepository
	Log    *zap.Logger
}

// NewIncidentCreatedConsumer creates a new IncidentCreatedConsumer.
func NewIncidentCreatedConsumer(client *redis.StreamClient, repo IncidentRepository, log *zap.Logger) *IncidentCreatedConsumer {
	return &IncidentCreatedConsumer{
		Client: client,
		Repo:   repo,
		Log:    log,
	}
}

// Start begins consuming incident.created events and processing them.
func (c *IncidentCreatedConsumer) Start(ctx context.Context, svc *court.Service) error {

	logger := c.Log.With(zap.String("component", "incident-created-consumer"))

	if err := c.Client.EnsureGroup(ctx); err != nil {
		logger.Error("failed to initialize stream consumer group", zap.Error(err))
		return err
	}

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var evt IncidentCreatedEvent

		if err := json.Unmarshal(data, &evt); err != nil {
			logger.Error("received invalid event payload", zap.Error(err))
			return err
		}

		if evt.IncidentID == "" {
			logger.Warn("received event without incident ID")
			return nil
		}

		logger.Info("processing new incident",
			zap.String("incident_id", evt.IncidentID),
		)

		incidentReport, err := c.Repo.GetByID(ctx, evt.IncidentID)
		if err != nil {
			logger.Error("unable to load incident data",
				zap.String("incident_id", evt.IncidentID),
				zap.Error(err),
			)
			return err
		}

		if err := svc.CreateTrial(ctx, incidentReport); err != nil {
			logger.Error("failed to create suit from incident",
				zap.String("incident_id", evt.IncidentID),
				zap.Error(err),
			)
			return err
		}

		logger.Info("suit successfully created from incident",
			zap.String("incident_id", evt.IncidentID),
		)

		return nil
	})
}
