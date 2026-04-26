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
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	goredis "github.com/redis/go-redis/v9"
)

// TestProcessMessage_Success verifies handler execution.
func TestProcessMessage_Success(t *testing.T) {
	called := false
	acked := false

	handler := func(ctx context.Context, data []byte) error {
		called = true

		if string(data) != "hello" {
			t.Errorf("unexpected payload: %s", data)
		}

		return nil
	}

	rdb := &testhelper.FakeRedisClient{
		XAckFunc: func(ctx context.Context, stream, group string, ids ...string) *goredis.IntCmd {
			acked = true
			return goredis.NewIntCmd(ctx)
		},
	}

	client := &StreamClient{
		rdb: rdb,
		cfg: Config{
			Stream: "s",
			Group:  "g",
		},
	}

	msg := goredis.XMessage{
		ID: "1-0",
		Values: map[string]any{
			"payload": "hello",
		},
	}

	client.processMessage(context.Background(), handler, msg)

	if !called {
		t.Fatal("expected handler to be called")
	}

	if !acked {
		t.Fatal("expected message to be acknowledged")
	}
}

// TestProcessMessage_HandlerError verifies no ack on handler failure.
func TestProcessMessage_HandlerError(t *testing.T) {
	handler := func(ctx context.Context, data []byte) error {
		return assertError{}
	}

	client := &StreamClient{}

	msg := goredis.XMessage{
		ID: "1-0",
		Values: map[string]any{
			"payload": "fail",
		},
	}

	// should not panic
	client.processMessage(context.Background(), handler, msg)
}

// TestProcessMessage_InvalidPayload verifies non-string payload handling.
func TestProcessMessage_InvalidPayload(t *testing.T) {
	client := &StreamClient{}

	handler := func(ctx context.Context, data []byte) error {
		t.Fatal("handler should not be called")
		return nil
	}

	msg := goredis.XMessage{
		ID: "1-0",
		Values: map[string]any{
			"payload": 123, // invalid type
		},
	}

	client.processMessage(context.Background(), handler, msg)
}

type assertError struct{}

func (assertError) Error() string { return "error" }
