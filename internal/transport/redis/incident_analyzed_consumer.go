package redisstream

import (
	"context"
	"encoding/json"

	"github.com/ghdrope/court/internal/court"
	"github.com/ghdrope/court/pkg/redis"
	"go.uber.org/zap"
)

// IncidentAnalyzedEvent is emitted after Prosecutor finishes analysis.
type IncidentAnalyzedEvent struct {
	IncidentID string `json:"incident_id"`
}

// IncidentAnalyzedConsumer consumes incident.analyzed events
// and creates Suit records.
type IncidentAnalyzedConsumer struct {
	Client *redis.StreamClient
	Log    *zap.Logger
}

// NewIncidentAnalyzedConsumer creates a new consumer.
func NewIncidentAnalyzedConsumer(client *redis.StreamClient, log *zap.Logger) *IncidentAnalyzedConsumer {
	return &IncidentAnalyzedConsumer{
		Client: client,
		Log:    log,
	}
}

// Start begins consuming incident.analyzed events.
func (c *IncidentAnalyzedConsumer) Start(ctx context.Context, svc *court.Service) error {

	logger := c.Log.With(zap.String("component", "court-consumer"))

	if err := c.Client.EnsureGroup(ctx); err != nil {
		return err
	}

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var evt IncidentAnalyzedEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			logger.Error("invalid event payload", zap.Error(err))
			return err
		}

		if evt.IncidentID == "" {
			logger.Error("missing incident_id in event payload")
			return nil
		}

		logger.Info("incident analyzed event received",
			zap.String("incident_id", evt.IncidentID),
		)

		if err := svc.CreateSuit(ctx, evt.IncidentID); err != nil {
			logger.Error("failed to create suit", zap.Error(err))
			return err
		}

		logger.Info("suit created successfully",
			zap.String("incident_id", evt.IncidentID),
		)

		return nil
	})
}
