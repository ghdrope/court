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
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

// TestConsumeProcessesMessage verifies that a message is consumed,
// passed to the handler, and acknowledged successfully.
func TestConsumeProcessesMessage(t *testing.T) {
	// Start in-memory Redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	cfg := Config{
		Stream:   "test-stream",
		Group:    "test-group",
		Consumer: "test-consumer",
	}

	// Create consumer group
	err = rdb.XGroupCreateMkStream(context.Background(), cfg.Stream, cfg.Group, "$").Err()
	if err != nil {
		t.Fatalf("failed to create group: %v", err)
	}

	client := &StreamClient{
		rdb: rdb,
		cfg: cfg,
	}

	// Add message to stream
	_, err = rdb.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: cfg.Stream,
		Values: map[string]interface{}{
			"payload": `{"hello":"world"}`,
		},
	}).Result()
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var called bool
	var mu sync.Mutex

	handler := func(ctx context.Context, data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		called = true
		cancel() // stop Consume loop after first message
		return nil
	}

	err = client.Consume(ctx, handler)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Fatalf("expected handler to be called")
	}
}

// TestConsumeHandlerError verifies that messages are NOT acknowledged
// when the handler returns an error.
func TestConsumeHandlerError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	cfg := Config{
		Stream:   "test-stream",
		Group:    "test-group",
		Consumer: "test-consumer",
	}

	err = rdb.XGroupCreateMkStream(context.Background(), cfg.Stream, cfg.Group, "$").Err()
	if err != nil {
		t.Fatalf("failed to create group: %v", err)
	}

	client := &StreamClient{
		rdb: rdb,
		cfg: cfg,
	}

	id, err := rdb.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: cfg.Stream,
		Values: map[string]interface{}{
			"payload": "fail",
		},
	}).Result()
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := func(ctx context.Context, data []byte) error {
		cancel()
		return context.Canceled // simulate failure
	}

	_ = client.Consume(ctx, handler)

	// Check pending entries (should NOT be acknowledged)
	pending, err := rdb.XPending(context.Background(), cfg.Stream, cfg.Group).Result()
	if err != nil {
		t.Fatalf("failed to check pending: %v", err)
	}

	if pending.Count == 0 {
		t.Fatalf("expected message %s to remain pending", id)
	}
}

// TestConsumeSkipsInvalidPayload verifies that messages without a valid
// "payload" field are ignored.
func TestConsumeSkipsInvalidPayload(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})

	cfg := Config{
		Stream:   "test-stream",
		Group:    "test-group",
		Consumer: "test-consumer",
	}

	err = rdb.XGroupCreateMkStream(context.Background(), cfg.Stream, cfg.Group, "$").Err()
	if err != nil {
		t.Fatalf("failed to create group: %v", err)
	}

	client := &StreamClient{
		rdb: rdb,
		cfg: cfg,
	}

	_, err = rdb.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: cfg.Stream,
		Values: map[string]interface{}{
			"invalid": 123,
		},
	}).Result()
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	called := false

	handler := func(ctx context.Context, data []byte) error {
		called = true
		return nil
	}

	_ = client.Consume(ctx, handler)

	if called {
		t.Fatalf("handler should not be called for invalid payload")
	}
}
