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

package officer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ghdrope/court/internal/incident"
	goredis "github.com/redis/go-redis/v9"
)

// HandleIncident persists an incident and emits a creation event.
//
// Behavior:
//   - Inserts or updates the incident in the repository
//   - Emits an event only when the incident is newly created
//
// This ensures idempotency and avoids duplicate event emission.
func (s *Service) HandleIncident(
	ctx context.Context,
	r *incident.IncidentReport,
) error {

	if r == nil {
		return fmt.Errorf("incident is nil")
	}

	logger := s.Log.WithValues(
		"incident_id", r.ID,
		"pod", r.Pod,
		"namespace", r.Namespace,
	)

	logger.Info("incident detected")

	created, err := s.IncidentRepo.Insert(ctx, r)
	if err != nil {
		logger.Error(err, "failed to store incident")
		return err
	}

	if created {
		logger.Info("incident stored")
	} else {
		logger.Info("incident updated")
	}

	// Emit event only for new incidents
	if !created {
		return nil
	}

	if s.RDB == nil {
		logger.Info("redis client is nil, skipping event emission")
		return nil
	}

	payload := map[string]any{
		"incident_id": r.ID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Error(err, "failed to encode incident event payload")
		return err
	}

	err = s.RDB.XAdd(ctx, &goredis.XAddArgs{
		Stream: IncidentCreatedStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err()

	if err != nil {
		logger.Error(err, "failed to publish incident event")
		return err
	}

	logger.Info("incident event published")

	return nil
}
