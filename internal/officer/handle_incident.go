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
	"go.uber.org/zap"
)

// HandleIncident persists the incident and emits an event.
//
// It ensures the incident is stored before publishing the event.
// This operation should be idempotent at the database level.
func (o *Service) HandleIncident(
	ctx context.Context,
	r *incident.IncidentReport,
) error {

	if r == nil {
		return fmt.Errorf("incident is nil")
	}

	logger := o.Log.With(
		zap.String("incident_id", r.ID),
		zap.String("cluster", r.Cluster),
		zap.String("namespace", r.Namespace),
		zap.String("pod", r.Pod),
	)

	logger.Info("handling incident")

	// Persist incident
	if err := o.Repo.Insert(ctx, r); err != nil {
		logger.Error("failed to store incident", zap.Error(err))
		return err
	}

	logger.Info("incident stored in archive")

	// Emit event with only the ID
	payload := map[string]any{
		"incident_id": r.ID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal event payload", zap.Error(err))
		return err
	}

	if err := o.RDB.XAdd(ctx, &goredis.XAddArgs{
		Stream: AnalyzedStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err(); err != nil {
		logger.Error("failed to publish event", zap.Error(err))
		return err
	}

	logger.Info("incident.created event published")

	return nil
}
