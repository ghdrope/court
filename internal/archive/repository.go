package archive

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	pb "github.com/ghdrope/court/proto/archive"
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

	containerIssues := req.ContainerIssues
	if containerIssues == nil {
		containerIssues = []*pb.ContainerIssue{}
	}

	ciJSON, err := json.Marshal(containerIssues)
	if err != nil {
		return err
	}

	logs := req.Logs
	if logs == nil {
		logs = []string{}
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
		req.EventId,
		req.PodName,
		req.Namespace,
		req.Phase,
		req.Reason,
		ciJSON,
		logs,
	)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	log.Printf("ROWS INSERTED = %d", rows)

	return err
}
