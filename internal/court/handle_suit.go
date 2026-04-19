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
package court

import (
	"context"
	"fmt"
	"time"

	"github.com/ghdrope/court/internal/suit"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateSuit creates a Suit from an analyzed incident.
func (s *Service) CreateSuit(ctx context.Context, incidentID string) error {

	if incidentID == "" {
		return fmt.Errorf("incidentID is empty")
	}

	log := s.Log.With(zap.String("incident_id", incidentID))

	// Check if suit already exists
	existing, err := s.Repo.GetByIncidentID(ctx, incidentID)
	if err == nil && existing != nil {
		log.Info("suit already exists, skipping creation")
		return nil
	}

	log.Info("creating suit")

	newSuit := &suit.Suit{
		ID:         uuid.NewString(),
		IncidentID: incidentID,
		Status:     suit.StatusOpen,
		CreatedAt:  time.Now(),
	}

	if err := s.Repo.Insert(ctx, newSuit); err != nil {
		log.Error("failed to insert suit", zap.Error(err))
		return err
	}

	log.Info("suit created",
		zap.String("suit_id", newSuit.ID),
	)

	return nil
}
