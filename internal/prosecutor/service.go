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

	"github.com/ghdrope/court/internal/incident"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Stream configuration for analyzed incidents.
const (
	AnalyzedStream = "incident.analyzed"
)

// IncidentRepository defines persistence contract.
type IncidentRepository interface {
	UpdateAnalysis(ctx context.Context, r *incident.IncidentReport) error
}

// Service handles post-processing over stored IncidentReports.
//
// It is responsible for orchestrating enrichment pipelines and
// persisting analysis results into the database.
type Service struct {
	Repo *incident.Repository
	RDB  *goredis.Client
	Log  *zap.Logger
}

// New creates a new Prosecutor service.
func New(repo *incident.Repository, rdb *goredis.Client, log *zap.Logger) *Service {
	return &Service{
		Repo: repo,
		RDB:  rdb,
		Log:  log,
	}
}
