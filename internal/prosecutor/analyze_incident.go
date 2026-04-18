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

package prosecutor

import (
	"context"
	"fmt"

	"github.com/ghdrope/court/internal/incident"
	"go.uber.org/zap"
)

// ProcessIncident performs analysis over an IncidentReport.
//
// It enriches the incident with:
//   - human-readable commentary
//   - optional repository reference URL
//
// The result is persisted through the repository layer.
// This operation must be idempotent.
func (s *Service) ProcessIncident(ctx context.Context, r *incident.IncidentReport) error {

	if r == nil {
		return fmt.Errorf("incident is nil")
	}

	logger := s.Log.With(
		zap.String("incident_id", r.ID),
	)

	logger.Info("starting incident analysis")

	// Simulated analysis (placeholder for LLM engine)
	commentary, repoURL := s.analyze(r)

	logger.Info("analysis completed",
		zap.String("commentary", commentary),
		zap.String("related_repo_url", repoURL),
	)

	// Attach analysis to domain model
	r.Analysis = &incident.ProsecutorAnalysis{
		Commentary:     commentary,
		RelatedRepoURL: repoURL,
	}

	// Persist result
	if err := s.Repo.UpdateAnalysis(ctx, r); err != nil {
		logger.Error("failed to update analysis", zap.Error(err))
		return err
	}

	logger.Info("analysis persisted to archive",
		zap.String("incident_id", r.ID),
	)

	return nil
}

// analysze contains the enrichment logic.
//
// This is intentionally isolated to make it easy to replace
// with LLM-based or rule-based engines later.
func (s *Service) analyze(r *incident.IncidentReport) (string, string) {
	// TBD
	return "nothing to add", "nothing to add"
}
