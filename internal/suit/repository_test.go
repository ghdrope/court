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
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ghdrope/court/pkg/testhelper"
)

// TestNewRepository verifies that a Repository is properly constructed.
func TestNewRepository(t *testing.T) {
	db, _ := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	if repo == nil {
		t.Fatal("expected repository, got nil")
	}

	if repo.DB != db {
		t.Fatal("repository DB not set correctly")
	}
}

// TestInitSchema verifies that the schema creation query is executed.
func TestInitSchema(t *testing.T) {
	db, mock := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS suits").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.InitSchema(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestInsert verifies that a valid Suit is inserted into the database.
func TestInsert(t *testing.T) {
	db, mock := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	now := time.Now()

	s := &Suit{
		ID:         "suit-1",
		IncidentID: "incident-1",
		Status:     StatusOpen,
		CreatedAt:  now,
		ClosedAt:   nil,
	}

	mock.ExpectExec("INSERT INTO suits").
		WithArgs(
			s.ID,
			s.IncidentID,
			s.Status,
			s.CreatedAt,
			s.ClosedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Insert(context.Background(), s)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestInsertNil verifies that inserting a nil Suit returns an error.
func TestInsertNil(t *testing.T) {
	db, _ := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	err := repo.Insert(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil suit")
	}
}

// TestGetByIncidentID verifies that a Suit is retrieved correctly.
func TestGetByIncidentID(t *testing.T) {
	db, mock := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"incident_id",
		"status",
		"created_at",
		"closed_at",
	}).AddRow(
		"suit-1",
		"incident-1",
		StatusOpen,
		now,
		nil,
	)

	mock.ExpectQuery("SELECT id, incident_id, status, created_at, closed_at").
		WithArgs("incident-1").
		WillReturnRows(rows)

	s, err := repo.GetByIncidentID(context.Background(), "incident-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if s == nil {
		t.Fatal("expected suit, got nil")
	}

	if s.ID != "suit-1" {
		t.Errorf("expected ID 'suit-1', got %s", s.ID)
	}

	if s.IncidentID != "incident-1" {
		t.Errorf("expected IncidentID 'incident-1', got %s", s.IncidentID)
	}
}

// TestGetByIncidentID_NotFound verifies that no result returns nil without error.
func TestGetByIncidentID_NotFound(t *testing.T) {
	db, mock := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	mock.ExpectQuery("SELECT id, incident_id, status, created_at, closed_at").
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	s, err := repo.GetByIncidentID(context.Background(), "missing")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if s != nil {
		t.Fatal("expected nil suit when not found")
	}
}

// TestClose verifies that a Suit is marked as closed.
func TestClose(t *testing.T) {
	db, mock := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	mock.ExpectExec("UPDATE suits").
		WithArgs(
			StatusClosed,
			sqlmock.AnyArg(), // time.Now()
			"suit-1",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Close(context.Background(), "suit-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
