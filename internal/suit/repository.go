package suit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository handles persistence of Suit entities.
type Repository struct {
	DB *sql.DB
}

// NewRepository creates a new Suit repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// InitSchema ensures suits table exists.
func (r *Repository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS suits (
		id TEXT PRIMARY KEY,
		incident_id TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		closed_at TIMESTAMPTZ
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_suits_incident_id_unique
	ON suits (incident_id);
	`

	_, err := r.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}

	return nil
}

// Insert creates a new Suit.
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
		closed_at
	)
	VALUES ($1,$2,$3,$4,$5)
	ON CONFLICT (incident_id) DO NOTHING
	`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		s.ID,
		s.IncidentID,
		s.Status,
		s.CreatedAt,
		s.ClosedAt,
	)

	if err != nil {
		return fmt.Errorf("insert suit: %w", err)
	}

	return nil
}

// GetByIncidentID retrieves a Suit by its incident reference.
func (r *Repository) GetByIncidentID(ctx context.Context, incidentID string) (*Suit, error) {
	query := `
	SELECT id, incident_id, status, created_at, closed_at
	FROM suits
	WHERE incident_id = $1
	`

	var s Suit

	err := r.DB.QueryRowContext(ctx, query, incidentID).Scan(
		&s.ID,
		&s.IncidentID,
		&s.Status,
		&s.CreatedAt,
		&s.ClosedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query suit: %w", err)
	}

	return &s, nil
}

// Close marks a Suit as closed.
func (r *Repository) Close(ctx context.Context, id string) error {
	now := time.Now()

	query := `
	UPDATE suits
	SET status = $1,
		closed_at = $2
	WHERE id = $3
	`

	_, err := r.DB.ExecContext(ctx, query, StatusClosed, now, id)
	if err != nil {
		return fmt.Errorf("close suit: %w", err)
	}

	return nil
}
