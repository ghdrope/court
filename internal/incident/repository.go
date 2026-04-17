package incident

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Repository handles persistence of IncidentReports.
//
// It is the only layer allowed to interact with the incidents table.
type Repository struct {
	DB *sql.DB
}

// NewRepository creates a new incident repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// InitSchema ensures the required database schema exists.
func (r *Repository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS incidents (
		id TEXT PRIMARY KEY,

		cluster TEXT NOT NULL,
		namespace TEXT NOT NULL,
		pod TEXT NOT NULL,

		events JSONB NOT NULL DEFAULT '[]',
		container_issues JSONB NOT NULL DEFAULT '[]',

		commentary TEXT NOT NULL DEFAULT '',
		related_repo_url TEXT NOT NULL DEFAULT '',

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	_, err := r.DB.ExecContext(ctx, query)
	return err
}

// Insert creates a new incident record in the database.
//
// This is called by the Officer when a new incident is detected.
func (r *Repository) Insert(ctx context.Context, inc *IncidentReport) error {
	if inc == nil {
		return fmt.Errorf("incident is nil")
	}

	eventsJSON, err := json.Marshal(inc.Events)
	if err != nil {
		return err
	}

	issuesJSON, err := json.Marshal(inc.ContainerIssues)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO incidents (
		id,
		cluster,
		namespace,
		pod,
		events,
		container_issues,
		commentary,
		related_repo_url,
		created_at,
		updated_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	ON CONFLICT (id) DO NOTHING
	`

	now := time.Now()

	_, err = r.DB.ExecContext(
		ctx,
		query,
		inc.ID,
		inc.Cluster,
		inc.Namespace,
		inc.Pod,
		eventsJSON,
		issuesJSON,
		"",
		"",
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("insert incident: %w", err)
	}

	return nil
}

// UpdateAnalysis updates the Prosecutor-generated analysis.
//
// This includes LLM commentary and optional repository reference.
func (r *Repository) UpdateAnalysis(ctx context.Context, inc *IncidentReport) error {
	if inc == nil || inc.Analysis == nil {
		return fmt.Errorf("invalid incident analysis")
	}

	query := `
	UPDATE incidents
	SET
		commentary = $1,
		related_repo_url = $2,
		updated_at = NOW()
	WHERE id = $3
	`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		inc.Analysis.Commentary,
		inc.Analysis.RelatedRepoURL,
		inc.ID,
	)

	if err != nil {
		return fmt.Errorf("update incident analysis: %w", err)
	}

	return nil
}
