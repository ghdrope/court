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

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/ghdrope/court/pkg/env"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultTestDSN = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"

// TestOpen verifies that sql.Open succeeds with a valid DSN.
func TestOpen(t *testing.T) {
	dsn := env.Get("DATABASE_URL", defaultTestDSN)

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if db == nil {
		t.Fatal("expected db to be non-nil")
	}

	_ = db.Close()
}

// TestPingWithRetryTimeout verifies timeout behavior when DB is unreachable.
func TestPingWithRetryTimeout(t *testing.T) {
	dsn := "postgres://postgres:postgres@localhost:59999/archive?sslmode=disable"

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = PingWithRetry(ctx, db)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestOpenConnectionSettings verifies connection pool configuration.
func TestOpenConnectionSettings(t *testing.T) {
	dsn := env.Get("DATABASE_URL", defaultTestDSN)

	db, err := Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	}()
	stats := db.Stats()

	if stats.MaxOpenConnections != 10 {
		t.Fatalf("expected max open connections to be 10, got %d", stats.MaxOpenConnections)
	}
}
