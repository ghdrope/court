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
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestInitSchema ensures schema creation runs without error.
func TestRepository_InitSchema(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS suits").
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo := NewRepository(db)

	if err := repo.InitSchema(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestInsertSuit ensures a suit is inserted correctly.
func TestRepository_Insert(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	now := time.Now()

	s := &Suit{
		ID:          "suit-1",
		IncidentID:  "incident-1",
		Status:      StatusOpen,
		CreatedAt:   now,
		ClosedAt:    nil,
		VCSIssueURL: "https://github.com/example/repo/issues/1",
	}

	mock.ExpectExec("INSERT INTO suits").
		WithArgs(
			s.ID,
			s.IncidentID,
			s.Status,
			s.CreatedAt,
			s.ClosedAt,
			s.VCSIssueURL,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewRepository(db)

	if err := repo.Insert(context.Background(), s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestGetByIncidentID ensures retrieval works correctly.
func TestRepository_GetByIncidentID(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "incident_id", "status", "created_at", "closed_at", "vcs_issue_url",
	}).AddRow(
		"suit-1",
		"incident-1",
		"open",
		now,
		nil,
		"https://github.com/example",
	)

	mock.ExpectQuery("SELECT (.+) FROM suits").
		WithArgs("incident-1").
		WillReturnRows(rows)

	repo := NewRepository(db)

	s, err := repo.GetByIncidentID(context.Background(), "incident-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s == nil {
		t.Fatalf("expected suit, got nil")
	}

	if s.ID != "suit-1" {
		t.Errorf("expected suit-1, got %s", s.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestListOpen ensures only open suits are returned.
func TestRepository_ListOpen(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "incident_id", "status", "created_at", "closed_at", "vcs_issue_url",
	}).AddRow(
		"suit-1",
		"incident-1",
		"open",
		now,
		nil,
		"https://github.com/example",
	)

	mock.ExpectQuery("SELECT (.+) FROM suits WHERE status").
		WithArgs(StatusOpen).
		WillReturnRows(rows)

	repo := NewRepository(db)

	result, err := repo.ListOpen(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	if result[0].Status != StatusOpen {
		t.Errorf("expected open status")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestClose ensures a suit is marked as closed.
func TestRepository_Close(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectExec("UPDATE suits").
		WithArgs(
			StatusClosed,
			sqlmock.AnyArg(),
			"suit-1",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRepository(db)

	if err := repo.Close(context.Background(), "suit-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpdateVCSInfo ensures VCS metadata is updated correctly.
func TestRepository_UpdateVCSInfo(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	now := time.Now()

	s := &Suit{
		ID:          "suit-1",
		IncidentID:  "incident-1",
		Status:      StatusOpen,
		CreatedAt:   now,
		ClosedAt:    nil,
		VCSIssueURL: "https://github.com/new",
	}

	mock.ExpectExec("UPDATE suits").
		WithArgs("https://github.com/new", "suit-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRepository(db)

	if err := repo.UpdateVCSInfo(context.Background(), s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
