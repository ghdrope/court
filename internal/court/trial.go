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

// CreateTrial is the main orchestration entrypoint for incident processing.
//
// It represents the lifecycle of a trial in the Court domain:
//
// Flow:
//  1. Ensure a Suit exists for the incident
//  2. Create a VCS issue describing the incident
//  3. Attach the issue metadata to the Suit
//  4. Persist the updated Suit state
func (s *Service) CreateTrial(
	ctx context.Context,
	inc *incident.IncidentReport,
) error {

	if inc == nil {
		return nil
	}

	// 1. Ensure suit exists for this incident
	suitEntity, err := s.CreateSuit(ctx, inc)
	if err != nil {
		return err
	}

	// 2. Create external VCS issue
	if err := s.CreateIssue(ctx, inc, suitEntity); err != nil {
		return err
	}

	// 3. Persist updated suit state
	return s.Repo.UpdateVCSInfo(ctx, suitEntity)
}

// CloseTrial closes an active Suit associated with an incident.
//
// The operation is idempotent:
//   - If no Suit exists → no-op
//   - If Suit is already closed → no-op
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

	logger.Info("closing suit")

	return s.Repo.Close(ctx, suitEntity.ID)
}
