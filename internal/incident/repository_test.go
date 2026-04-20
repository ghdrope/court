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
	"testing"

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

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS incidents").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.InitSchema(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestInsert verifies that a valid incident is inserted into the database.
func TestInsert(t *testing.T) {
	db, mock := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	inc := &IncidentReport{
		ID:              "test-id",
		Cluster:         "cluster-1",
		Namespace:       "default",
		Pod:             "pod-1",
		Events:          []K8sEvent{},
		ContainerIssues: []ContainerIssue{},
	}

	mock.ExpectExec("INSERT INTO incidents").
		WithArgs(
			inc.ID,
			inc.Cluster,
			inc.Namespace,
			inc.Pod,
			sqlmock.AnyArg(), // events JSON
			sqlmock.AnyArg(), // issues JSON
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Insert(context.Background(), inc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestInsertNil verifies that inserting a nil incident returns an error.
func TestInsertNil(t *testing.T) {
	db, _ := testhelper.NewTestDB(t)
	defer testhelper.CloseDB(t, db)

	repo := NewRepository(db)

	err := repo.Insert(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil incident")
	}
}
