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
