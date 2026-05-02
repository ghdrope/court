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

package incident

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRepository_Insert ensures incidents are inserted or updated correctly.
func TestRepository_Insert(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	inc := &IncidentReport{
		ID:         "ns-1/api/uid-123",
		Cluster:    "cluster-1",
		Namespace:  "ns-1",
		Pod:        "api",
		VCSRepoURL: "https://github.com/example/repo",
	}

	mock.ExpectQuery("INSERT INTO incidents").
		WillReturnRows(sqlmock.NewRows([]string{"inserted"}).AddRow(true))

	repo := NewRepository(db)

	created, err := repo.Insert(context.Background(), inc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !created {
		t.Errorf("expected created=true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestRepository_GetByID ensures retrieval and JSON decoding works correctly.
func TestRepository_GetByID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectQuery("SELECT").
		WithArgs("ns-1/api/uid-123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"cluster",
			"namespace",
			"pod",
			"github_repo_url",
			"events",
			"containers_metadata",
		}).AddRow(
			"ns-1/api/uid-123",
			"cluster-1",
			"ns-1",
			"api",
			"https://github.com/example",
			`[]`,
			`[]`,
		))

	repo := NewRepository(db)

	inc, err := repo.GetByID(context.Background(), "ns-1/api/uid-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inc.Pod != "api" {
		t.Errorf("expected api, got %s", inc.Pod)
	}
}

// TestRepository_GetByID_NotFound ensures missing records return ErrNotFound.
func TestRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	repo := NewRepository(db)

	_, err = repo.GetByID(context.Background(), "missing")

	if err == nil {
		t.Errorf("expected error for missing incident")
	}

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
