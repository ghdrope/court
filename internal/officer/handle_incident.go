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

// HandleIncident stores the incident and emits an event.
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

	if err := s.IncidentRepo.Insert(ctx, r); err != nil {
		logger.Error(err, "failed to store incident")
		return err
	}

	logger.Info("incident stored")

	// Emit event with only the ID
	payload := map[string]any{
		"incident_id": r.ID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Error(err, "failed to encode event payload")
		return err
	}

	err = s.RDB.XAdd(ctx, &goredis.XAddArgs{
		Stream: IncidentCreatedStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err()

	if err != nil {
		logger.Error(err, "failed to publish event")
		return err
	}

	logger.Info("incident event published")

	return nil
}
