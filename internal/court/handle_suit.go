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

// CreateSuit ensures a Suit exists for the given incident.
func (s *Service) CreateSuit(
	ctx context.Context,
	inc *incident.IncidentReport,
) (*suit.Suit, error) {

	if inc == nil {
		return nil, fmt.Errorf("incident is nil")
	}

	log := s.Log.With(zap.String("incident_id", inc.ID))

	existing, err := s.Repo.GetByIncidentID(ctx, inc.ID)
	if err == nil && existing != nil {
		log.Info("suit already exists")
		return existing, nil
	}

	log.Info("creating suit for incident")

	newSuit := &suit.Suit{
		ID:         uuid.NewString(),
		IncidentID: inc.ID,
		Status:     suit.StatusOpen,
		CreatedAt:  time.Now(),
	}

	if err := s.Repo.Insert(ctx, newSuit); err != nil {
		log.Error("failed to create suit",
			zap.Error(err),
		)
		return nil, err
	}

	log.Info("suit created successfully",
		zap.String("suit_id", newSuit.ID),
	)

	return newSuit, nil
}
