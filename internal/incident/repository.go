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

package incident

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// Repository handles persistence of IncidentReport entities.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new incident repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ErrNotFound is returned when an incident does not exist.
var ErrNotFound = errors.New("incident not found")

// InitSchema ensures that the incidents table exists.
func (r *Repository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS incidents (
		id TEXT PRIMARY KEY,

		cluster TEXT NOT NULL,
		namespace TEXT NOT NULL,
		pod TEXT NOT NULL,

		github_repo_url TEXT,

		events JSONB NOT NULL DEFAULT '[]',
		container_issues JSONB NOT NULL DEFAULT '[]',

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_incidents_namespace_pod
	ON incidents (namespace, pod);
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("init incident schema: %w", err)
	}

	return nil
}

// Insert creates or updates an IncidentReport.
func (r *Repository) Insert(ctx context.Context, inc *IncidentReport) (bool, error) {
	if inc == nil {
		return false, fmt.Errorf("incident is nil")
	}

	eventsJSON, err := json.Marshal(inc.Events)
	if err != nil {
		return false, err
	}

	issuesJSON, err := json.Marshal(inc.ContainerIssues)
	if err != nil {
		return false, err
	}

	query := `
	INSERT INTO incidents (
		id,
		cluster,
		namespace,
		pod,
		github_repo_url,
		events,
		container_issues
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	ON CONFLICT (id) DO UPDATE SET
		cluster = EXCLUDED.cluster,
		namespace = EXCLUDED.namespace,
		pod = EXCLUDED.pod,
		github_repo_url = EXCLUDED.github_repo_url,
		events = EXCLUDED.events,
		container_issues = EXCLUDED.container_issues,
		updated_at = NOW()
	RETURNING (xmax = 0) AS inserted
	`

	var created bool

	err = r.db.QueryRowContext(
		ctx,
		query,
		inc.ID,
		inc.Cluster,
		inc.Namespace,
		inc.Pod,
		inc.VCSRepoURL,
		eventsJSON,
		issuesJSON,
	).Scan(&created)

	if err != nil {
		return false, fmt.Errorf("insert incident: %w", err)
	}

	return created, nil
}

// GetByID retrieves an IncidentReport by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*IncidentReport, error) {
	query := `
	SELECT 
		id,
		cluster,
		namespace,
		pod,
		github_repo_url,
		events,
		container_issues
	FROM incidents
	WHERE id = $1
	`

	var (
		inc        IncidentReport
		eventsJSON []byte
		issuesJSON []byte
	)

	err := r.db.QueryRowContext(ctx, query, id).
		Scan(
			&inc.ID,
			&inc.Cluster,
			&inc.Namespace,
			&inc.Pod,
			&inc.VCSRepoURL,
			&eventsJSON,
			&issuesJSON,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query incident: %w", err)
	}

	if len(eventsJSON) > 0 {
		if err := json.Unmarshal(eventsJSON, &inc.Events); err != nil {
			return nil, fmt.Errorf("unmarshal events: %w", err)
		}
	}

	if len(issuesJSON) > 0 {
		if err := json.Unmarshal(issuesJSON, &inc.ContainerIssues); err != nil {
			return nil, fmt.Errorf("unmarshal container issues: %w", err)
		}
	}

	return &inc, nil
}
