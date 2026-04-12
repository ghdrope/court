package archive

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	pb "github.com/ghdrope/court/proto/archive"
	"github.com/lib/pq"
)

// TestPostgresRepository_Store_Success validates successful persistence of an incident.
func TestPostgresRepository_Store_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	repo := NewPostgresRepository(db)

	req := &pb.StoreIncidentRequest{
		Id:        "evt-123",
		PodName:   "pod-1",
		Namespace: "default",
		Phase:     "Failed",
		Reason:    "CrashLoopBackOff",
		ContainerIssues: []*pb.ContainerIssue{
			{Container: "app"},
		},
		Logs: []string{"log1", "log2"},
	}

	expectedJSON, _ := json.Marshal(req.ContainerIssues)

	mock.ExpectExec(regexp.QuoteMeta(`
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
	`)).
		WithArgs(
			req.Id,
			req.PodName,
			req.Namespace,
			req.Phase,
			req.Reason,
			expectedJSON,
			pq.StringArray(req.Logs),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Store(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPostgresRepository_Store_NilRequest tests nil request
func TestPostgresRepository_Store_NilRequest(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	repo := NewPostgresRepository(db)

	err := repo.Store(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

// TestPostgresRepository_Store_DBError ensures DB execution errors are wrapped properly.
func TestPostgresRepository_Store_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	repo := NewPostgresRepository(db)

	req := &pb.StoreIncidentRequest{
		Id:        "evt-123",
		PodName:   "pod-1",
		Namespace: "default",
		Phase:     "Failed",
		Reason:    "CrashLoopBackOff",
	}

	mock.ExpectExec("INSERT INTO incidents").
		WillReturnError(errors.New("db failure"))

	err := repo.Store(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() == "" {
		t.Fatal("expected wrapped error message")
	}
}

// TestPostgresRepository_Store_DefaultValues tests default normalization
func TestPostgresRepository_Store_DefaultValues(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	repo := NewPostgresRepository(db)

	req := &pb.StoreIncidentRequest{
		Id:        "evt-999",
		PodName:   "pod-x",
		Namespace: "default",
		Phase:     "Running",
		Reason:    "OK",
	}

	mock.ExpectExec("INSERT INTO incidents").
		WithArgs(
			req.Id,
			req.PodName,
			req.Namespace,
			req.Phase,
			req.Reason,
			sqlmock.AnyArg(),
			pq.StringArray{},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Store(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPostgresRepository_Store_RowsAffectedError simulates RowsAffected failures.
func TestPostgresRepository_Store_RowsAffectedError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	repo := NewPostgresRepository(db)

	req := &pb.StoreIncidentRequest{
		Id:        "evt-1",
		PodName:   "pod",
		Namespace: "ns",
		Phase:     "Failed",
		Reason:    "error",
	}

	result := sqlmock.NewErrorResult(errors.New("rows error"))

	mock.ExpectExec("INSERT INTO incidents").
		WillReturnResult(result)

	err := repo.Store(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
