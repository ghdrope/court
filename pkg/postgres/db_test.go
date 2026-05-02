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
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestDefaultConfig verifies that DefaultConfig returns expected defaults.
func TestDefaultConfig(t *testing.T) {
	dsn := "postgres://user:pass@localhost:5432/db"

	cfg := DefaultConfig(dsn)

	if cfg.DSN != dsn {
		t.Errorf("unexpected DSN: got %q, want %q", cfg.DSN, dsn)
	}

	if cfg.MaxOpenConns != 10 {
		t.Errorf("unexpected MaxOpenConns: got %d, want 10", cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns != 5 {
		t.Errorf("unexpected MaxIdleConns: got %d, want 5", cfg.MaxIdleConns)
	}

	if cfg.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("unexpected ConnMaxLifetime")
	}

	if cfg.ConnMaxIdleTime != 5*time.Minute {
		t.Errorf("unexpected ConnMaxIdleTime")
	}
}

// TestOpen verifies that Open initializes a DB without error.
func TestOpen(t *testing.T) {
	cfg := DefaultConfig("postgres://invalid") // no real connection needed

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db == nil || db.DB == nil {
		t.Fatal("expected non-nil DB")
	}
}
