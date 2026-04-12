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
	"log"

	pb "github.com/ghdrope/court/proto/archive"
	"github.com/lib/pq"
)

// Repository defines persistence operations for incidents.
type Repository interface {
	Store(ctx context.Context, req *pb.StoreIncidentRequest) error
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	DB *sql.DB
}

// NewPostgresRepository creates a new repository instance.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

// Store persists an incident into PostgreSQL.
func (r *PostgresRepository) Store(
	ctx context.Context,
	req *pb.StoreIncidentRequest,
) error {

	if req == nil {
		return fmt.Errorf("store incident: request is nil")
	}

	containerIssues := req.ContainerIssues
	if containerIssues == nil {
		containerIssues = []*pb.ContainerIssue{}
	}

	ciJSON, err := json.Marshal(containerIssues)
	if err != nil {
		return fmt.Errorf("marshal container issues: %w", err)
	}

	logs := pq.StringArray(req.Logs)
	if req.Logs == nil {
		logs = pq.StringArray{}
	}

	query := `
		INSERT INTO incidents (
			event_id,
			pod_name,
			namespace,
			phase,
			reason,
			container_issues,
			logs
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`

	res, err := r.DB.ExecContext(
		ctx,
		query,
		req.Id,
		req.PodName,
		req.Namespace,
		req.Phase,
		req.Reason,
		ciJSON,
		logs,
	)
	if err != nil {
		return fmt.Errorf("insert incident: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	log.Printf("incident stored id=%s rows=%d", req.Id, rows)

	return nil
}
