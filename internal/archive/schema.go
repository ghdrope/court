package archive

import (
	"context"
	"database/sql"
)

// InitSchema ensures the required database schema exists.
func InitSchema(ctx context.Context, db *sql.DB) error {

	query := `
	CREATE TABLE IF NOT EXISTS incidents (
		id SERIAL PRIMARY KEY,

		event_id TEXT NOT NULL UNIQUE,
		pod_name TEXT NOT NULL,
		namespace TEXT NOT NULL,

		phase TEXT NOT NULL,
		reason TEXT NOT NULL,

		container_issues JSONB NOT NULL DEFAULT '[]',
		logs TEXT[] NOT NULL DEFAULT '{}',

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_incidents_event_id ON incidents(event_id);
	CREATE INDEX IF NOT EXISTS idx_incidents_namespace ON incidents(namespace);
	`

	_, err := db.ExecContext(ctx, query)
	return err
}
