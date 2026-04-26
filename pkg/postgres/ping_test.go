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

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestPingWithRetry_ContextCancel verifies that PingWithRetry respects context cancellation.
func TestPingWithRetry_ContextCancel(t *testing.T) {
	cfg := DefaultConfig("postgres://invalid")
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err = db.PingWithRetry(ctx)

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}

	if err != context.Canceled {
		t.Errorf("unexpected error: got %v, want context.Canceled", err)
	}
}

// TestPingWithRetry_Timeout verifies that timeout is reached.
func TestPingWithRetry_Timeout(t *testing.T) {
	cfg := DefaultConfig("postgres://invalid")
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	start := time.Now()
	err = db.PingWithRetry(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	if elapsed < 30*time.Second {
		t.Errorf("expected retry duration close to timeout, got %v", elapsed)
	}
}
