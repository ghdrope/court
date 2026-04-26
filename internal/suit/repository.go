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

package suit

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Repository provides database access for Suit persistence operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Suit repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// InitSchema ensures that the required database schema exists.
func (r *Repository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS suits (
		id TEXT PRIMARY KEY,
		incident_id TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		closed_at TIMESTAMPTZ,
		vcs_issue_url TEXT
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_suits_incident_id_unique
	ON suits (incident_id);
	`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("init suit schema: %w", err)
	}

	return nil
}

// Insert creates a new Suit.
// If a Suit already exists for the same incident_id, the insert is ignored (no update is performed).
func (r *Repository) Insert(ctx context.Context, s *Suit) error {
	if s == nil {
		return fmt.Errorf("suit is nil")
	}

	query := `
	INSERT INTO suits (
		id,
		incident_id,
		status,
		created_at,
		closed_at,
		vcs_issue_url
	)
	VALUES ($1,$2,$3,$4,$5,$6)
	ON CONFLICT (incident_id) DO NOTHING
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		s.ID,
		s.IncidentID,
		s.Status,
		s.CreatedAt,
		s.ClosedAt,
		s.VCSIssueURL,
	)

	if err != nil {
		return fmt.Errorf("insert suit: %w", err)
	}

	return nil
}

// UpdateVCSInfo updates external VCS metadata for a Suit.
//
// This is typically called after creating an external issue
// and storing its URL.
func (r *Repository) UpdateVCSInfo(ctx context.Context, s *Suit) error {
	if s == nil {
		return fmt.Errorf("suit is nil")
	}

	query := `
	UPDATE suits
	SET vcs_issue_url = $1
	WHERE id = $2
	`

	if _, err := r.db.ExecContext(ctx, query, s.VCSIssueURL, s.ID); err != nil {
		return fmt.Errorf("update vcs info: %w", err)
	}

	return nil
}

// GetByIncidentID retrieves a Suit by its incident ID.
// Returns (nil, nil) if no Suit is found.
func (r *Repository) GetByIncidentID(ctx context.Context, incidentID string) (*Suit, error) {
	query := `
	SELECT id, incident_id, status, created_at, closed_at, vcs_issue_url
	FROM suits
	WHERE incident_id = $1
	`

	var s Suit

	err := r.db.QueryRowContext(ctx, query, incidentID).Scan(
		&s.ID,
		&s.IncidentID,
		&s.Status,
		&s.CreatedAt,
		&s.ClosedAt,
		&s.VCSIssueURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get suit by incident id: %w", err)
	}

	return &s, nil
}

// ListOpen returns all suits that are currently open.
func (r *Repository) ListOpen(ctx context.Context) ([]Suit, error) {
	query := `
	SELECT id, incident_id, status, created_at, closed_at, vcs_issue_url
	FROM suits
	WHERE status = $1
	`

	rows, err := r.db.QueryContext(ctx, query, StatusOpen)
	if err != nil {
		return nil, fmt.Errorf("list open suits: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("error closing rows: %v", cerr)
		}
	}()

	var result []Suit

	for rows.Next() {
		var s Suit

		if err := rows.Scan(
			&s.ID,
			&s.IncidentID,
			&s.Status,
			&s.CreatedAt,
			&s.ClosedAt,
			&s.VCSIssueURL,
		); err != nil {
			return nil, fmt.Errorf("scan suit row: %w", err)
		}

		result = append(result, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate suit rows: %w", err)
	}

	return result, nil
}

// Close marks a Suit as closed by updating its status and setting closed_at to current time.
func (r *Repository) Close(ctx context.Context, id string) error {
	now := time.Now()

	query := `
	UPDATE suits
	SET status = $1,
		closed_at = $2
	WHERE id = $3
	`

	if _, err := r.db.ExecContext(ctx, query, StatusClosed, now, id); err != nil {
		return fmt.Errorf("close suit: %w", err)
	}

	return nil
}
