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
	"encoding/json"
	"fmt"
	"time"

	"github.com/ghdrope/court/internal/incident"
)

// initIncidentSchema creates the incidents table if it does not exist.
func initIncidentSchema(ctx context.Context, db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS incidents (
		id SERIAL PRIMARY KEY,

		event_id TEXT NOT NULL UNIQUE,
		cluster TEXT NOT NULL,
		namespace TEXT NOT NULL,
		pod TEXT NOT NULL,

		events JSONB NOT NULL DEFAULT '[]',
		container_issues JSONB NOT NULL DEFAULT '[]',

		prosecutor_commentary TEXT NOT NULL DEFAULT '',

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_incidents_event_id ON incidents(event_id);
	CREATE INDEX IF NOT EXISTS idx_incidents_namespace ON incidents(namespace);
	`

	_, err := db.ExecContext(ctx, query)
	return err
}

// StoreIncident persists an IncidentReport into the Archive
// and emits a generic "stored" event.
func (a *Archive) StoreIncident(
	ctx context.Context,
	r *incident.IncidentReport,
) error {

	if r == nil {
		return fmt.Errorf("incident is nil")
	}

	eventsJSON, err := json.Marshal(r.Events)
	if err != nil {
		return fmt.Errorf("marshal events: %w", err)
	}

	containerIssuesJSON, err := json.Marshal(r.ContainerIssues)
	if err != nil {
		return fmt.Errorf("marshal container issues: %w", err)
	}

	query := `
		INSERT INTO incidents (
			event_id,
			cluster,
			namespace,
			pod,
			events,
			container_issues,
			prosecutor_commentary,
			created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`

	_, err = a.DB.ExecContext(
		ctx,
		query,
		r.ID,
		r.Cluster,
		r.Namespace,
		r.Pod,
		eventsJSON,
		containerIssuesJSON,
		r.ProsecutorCommentary,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("insert incident: %w", err)
	}

	// Emit generic stored event
	if a.Publisher != nil {
		event := StoredEvent{
			Type:    "incident.stored",
			Payload: r,
		}

		_ = a.Publisher.PublishStored(ctx, event)
	}

	return nil
}
