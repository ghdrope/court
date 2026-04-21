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

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/suit"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// IncidentCreatedStream configuration for created incidents.
const (
	IncidentCreatedStream = "incident.created"
)

// IncidentRepository defines persistence for incidents.
type IncidentRepository interface {
	Insert(ctx context.Context, r *incident.IncidentReport) error
}

// SuitRepository defines persistence for suits.
type SuitRepository interface {
	ListOpen(ctx context.Context) ([]suit.Suit, error)
	Close(ctx context.Context, id string) error
}

// Service handles incident creation logic and suit lifecycle operations.
//
// Responsibilities:
//   - Persist IncidentReports
//   - Emit events to Redis
//   - Recover open suits after restarts
type Service struct {
	IncidentRepo IncidentRepository
	SuitRepo     SuitRepository
	RDB          *goredis.Client
	Log          *zap.Logger
}

// New creates a new Officer service.
func New(incidentRepo IncidentRepository, suitRepo SuitRepository, rdb *goredis.Client, logger *zap.Logger) *Service {
	return &Service{
		IncidentRepo: incidentRepo,
		SuitRepo:     suitRepo,
		RDB:          rdb,
		Log:          logger,
	}
}
