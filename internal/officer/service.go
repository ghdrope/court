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
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Stream configuration for created incidents.
const (
	AnalyzedStream = "incident.created"
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
