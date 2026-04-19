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

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/suit"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateSuit creates a Suit from an analyzed incident and
// publishes a GitHub issue.
func (s *Service) CreateSuit(ctx context.Context, inc *incident.IncidentReport) error {

	if inc == nil {
		return fmt.Errorf("incident is nil")
	}

	log := s.Log.With(zap.String("incident_id", inc.ID))

	// Check if suit already exists
	existing, err := s.Repo.GetByIncidentID(ctx, inc.ID)
	if err == nil && existing != nil {
		log.Info("suit already exists, skipping creation")
		return nil
	}

	log.Info("creating suit")

	newSuit := &suit.Suit{
		ID:         uuid.NewString(),
		IncidentID: inc.ID,
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

	if s.GitHub != nil {
		if err := s.createGitHubIssue(ctx, inc); err != nil {
			log.Error("failed to create github issue", zap.Error(err))
		}
	}

	return nil
}
