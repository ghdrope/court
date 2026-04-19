package officer

import (
	"context"

	"github.com/ghdrope/court/internal/incident"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// IncidentRepository defines the minimal contract required by the service.
type IncidentRepository interface {
	Insert(ctx context.Context, r *incident.IncidentReport) error
}

// Service handles incident creation logic.
//
// It is responsible for persisting IncidentReports and emitting
// upstream events for further processing.
type Service struct {
	Repo IncidentRepository
	RDB  *goredis.Client
	Log  *zap.Logger
}

// New creates a new Officer service.
func New(repo IncidentRepository, rdb *goredis.Client, logger *zap.Logger) *Service {
	return &Service{
		Repo: repo,
		RDB:  rdb,
		Log:  logger,
	}
}
