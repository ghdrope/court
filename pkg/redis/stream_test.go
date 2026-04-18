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

package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

// TestEnsureGroupCreatesGroup verifies that EnsureGroup
// successfully creates a consumer group when it does not exist.
func TestEnsureGroupCreatesGroup(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	cfg := Config{
		Stream: "test-stream",
		Group:  "test-group",
	}

	client := NewStreamClient(rdb, cfg)

	err = client.EnsureGroup(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify group exists by querying XINFO GROUPS
	groups, err := rdb.XInfoGroups(context.Background(), cfg.Stream).Result()
	if err != nil {
		t.Fatalf("failed to fetch groups: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	if groups[0].Name != cfg.Group {
		t.Fatalf("expected group name %q, got %q", cfg.Group, groups[0].Name)
	}
}

// TestEnsureGroupIsIdempotent verifies that calling EnsureGroup
// multiple times does not return an error.
func TestEnsureGroupIsIdempotent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	cfg := Config{
		Stream: "test-stream",
		Group:  "test-group",
	}

	client := NewStreamClient(rdb, cfg)

	// First call
	if err := client.EnsureGroup(context.Background()); err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}

	// Second call (should NOT fail)
	if err := client.EnsureGroup(context.Background()); err != nil {
		t.Fatalf("expected idempotent behavior, got error: %v", err)
	}
}

// TestIsBusyGroupError verifies detection of BUSYGROUP errors.
func TestIsBusyGroupError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non BUSYGROUP error",
			err:  errors.New("some random error"),
			want: false,
		},
		{
			name: "BUSYGROUP error",
			err:  errors.New("BUSYGROUP Consumer Group name already exists"),
			want: true,
		},
		{
			name: "BUSYGROUP substring",
			err:  errors.New("ERR BUSYGROUP something"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBusyGroupError(tt.err)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
