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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestInitSchema_Success ensures schema initialization executes correctly.
func TestInitSchema_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS incidents").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = InitSchema(context.Background(), db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestInitSchema_DBError ensures schema errors are propagated.
func TestInitSchema_DBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS incidents").
		WillReturnError(sql.ErrConnDone)

	err := InitSchema(context.Background(), db)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestInitSchema_ContextCancellation simulates context cancellation.
func TestInitSchema_ContextCancellation(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() {
		_ = db.Close()
	}()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS incidents").
		WillDelayFor(0) // immediate

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := InitSchema(ctx, db)
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}
