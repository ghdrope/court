package officer

import (
	"github.com/ghdrope/court/internal/incident"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Service handles incident creation logic.
//
// It is responsible for persisting IncidentReports and emitting
// upstream events for further processing.
type Service struct {
	Repo *incident.Repository
	RDB  *goredis.Client
	Log  *zap.Logger
}

// New creates a new Officer service.
func New(repo *incident.Repository, rdb *goredis.Client, logger *zap.Logger) *Service {
	return &Service{
		Repo: repo,
		RDB:  rdb,
		Log:  logger,
	}
}
