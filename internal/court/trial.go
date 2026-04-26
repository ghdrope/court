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

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/suit"
	"go.uber.org/zap"
)

// CreateTrial orchestrates the full incident lifecycle.
//
// Flow:
//  1. Ensure Suit exists for the incident
//  2. Create VCS Issue describing the incident
//  3. Attach Issue metadata to the Suit
//  4. Persist updated Suit state
func (s *Service) CreateTrial(
	ctx context.Context,
	inc *incident.IncidentReport,
) error {

	if inc == nil {
		return nil
	}

	suitEntity, err := s.CreateSuit(ctx, inc)
	if err != nil {
		return err
	}

	if err := s.CreateIssue(ctx, inc, suitEntity); err != nil {
		return err
	}

	return s.Repo.UpdateVCSInfo(ctx, suitEntity)
}

// CloseTrial closes an active Suit associated with an incident.
//
// The operation is idempotent:
//   - If no Suit exists → no-op
//   - If already closed → no-op
func (s *Service) CloseTrial(
	ctx context.Context,
	incidentID string,
	reason string,
) error {

	logger := s.Log.With(
		zap.String("incident_id", incidentID),
		zap.String("reason", reason),
	)

	suitEntity, err := s.Repo.GetByIncidentID(ctx, incidentID)
	if err != nil {
		return err
	}

	if suitEntity == nil {
		logger.Info("no suit found, nothing to close")
		return nil
	}

	if suitEntity.Status == suit.StatusClosed {
		logger.Info("suit already closed",
			zap.String("suit_id", suitEntity.ID),
		)
		return nil
	}

	logger.Info("closing trial")

	// Close issue
	if err := s.CloseIssue(ctx, suitEntity); err != nil {
		logger.Error("failed to close VCS issue",
			zap.String("suit_id", suitEntity.ID),
			zap.Error(err),
		)
		return err
	}

	// Persist updated state
	if err := s.Repo.Close(ctx, suitEntity.ID); err != nil {
		logger.Error("failed to persist suit close",
			zap.String("suit_id", suitEntity.ID),
			zap.Error(err),
		)
		return err
	}

	logger.Info("trial closed successfully",
		zap.String("suit_id", suitEntity.ID),
	)

	return nil
}
