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

	"github.com/ghdrope/court/internal/court"
	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/pkg/redis"
	"go.uber.org/zap"
)

// IncidentAnalyzedEvent is emitted after Prosecutor finishes analysis.
type IncidentAnalyzedEvent struct {
	IncidentID string `json:"incident_id"`
}

// IncidentRepository defines minimal read access needed by the consumer.
type IncidentRepository interface {
	GetByID(ctx context.Context, id string) (*incident.IncidentReport, error)
}

// IncidentAnalyzedConsumer consumes incident.analyzed events
// and creates Suit records.
type IncidentAnalyzedConsumer struct {
	Client *redis.StreamClient
	Repo   IncidentRepository
	Log    *zap.Logger
}

// NewIncidentAnalyzedConsumer creates a new consumer.
func NewIncidentAnalyzedConsumer(client *redis.StreamClient, repo IncidentRepository, log *zap.Logger) *IncidentAnalyzedConsumer {
	return &IncidentAnalyzedConsumer{
		Client: client,
		Repo:   repo,
		Log:    log,
	}
}

// Start begins consuming incident.analyzed events.
func (c *IncidentAnalyzedConsumer) Start(ctx context.Context, svc *court.Service) error {

	logger := c.Log.With(zap.String("component", "court-consumer"))

	if err := c.Client.EnsureGroup(ctx); err != nil {
		return err
	}

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var evt IncidentAnalyzedEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			logger.Error("invalid event payload", zap.Error(err))
			return err
		}

		if evt.IncidentID == "" {
			logger.Error("missing incident_id in event payload")
			return nil
		}

		logger.Info("incident analyzed event received",
			zap.String("incident_id", evt.IncidentID),
		)

		inc, err := c.Repo.GetByID(ctx, evt.IncidentID)
		if err != nil {
			logger.Error("failed to load incident",
				zap.String("incident_id", evt.IncidentID),
				zap.Error(err),
			)
			return fmt.Errorf("load incident: %w", err)
		}

		if err := svc.CreateSuit(ctx, inc); err != nil {
			logger.Error("failed to create suit",
				zap.String("incident_id", evt.IncidentID),
				zap.Error(err),
			)
			return err
		}

		logger.Info("suit created successfully",
			zap.String("incident_id", evt.IncidentID),
		)

		return nil
	})
}
