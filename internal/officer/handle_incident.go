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
