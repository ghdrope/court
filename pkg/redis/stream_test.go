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

	"github.com/ghdrope/court/internal/testhelper"
	goredis "github.com/redis/go-redis/v9"
)

// TestEnsureGroup_Success verifies that EnsureGroup succeeds when no error is returned.
func TestEnsureGroup_Success(t *testing.T) {
	rdb := &testhelper.FakeRedisClient{
		XGroupCreateMkStreamFunc: func(ctx context.Context, stream, group, start string) *goredis.StatusCmd {
			cmd := goredis.NewStatusCmd(ctx)
			cmd.SetErr(nil)
			return cmd
		},
	}

	client := &StreamClient{
		rdb: rdb,
		cfg: Config{
			Stream: "stream",
			Group:  "group",
		},
	}

	if err := client.EnsureGroup(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestEnsureGroup_BusyGroup verifies that BUSYGROUP errors are ignored.
func TestEnsureGroup_BusyGroup(t *testing.T) {
	rdb := &testhelper.FakeRedisClient{
		XGroupCreateMkStreamFunc: func(ctx context.Context, stream, group, start string) *goredis.StatusCmd {
			cmd := goredis.NewStatusCmd(ctx)
			cmd.SetErr(errors.New("BUSYGROUP Consumer Group name already exists"))
			return cmd
		},
	}

	client := &StreamClient{
		rdb: rdb,
		cfg: Config{
			Stream: "stream",
			Group:  "group",
		},
	}

	if err := client.EnsureGroup(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestEnsureGroup_Error verifies that unexpected errors are returned.
func TestEnsureGroup_Error(t *testing.T) {
	rdb := &testhelper.FakeRedisClient{
		XGroupCreateMkStreamFunc: func(ctx context.Context, stream, group, start string) *goredis.StatusCmd {
			cmd := goredis.NewStatusCmd(ctx)
			cmd.SetErr(errors.New("boom"))
			return cmd
		},
	}

	client := &StreamClient{
		rdb: rdb,
		cfg: Config{
			Stream: "stream",
			Group:  "group",
		},
	}

	if err := client.EnsureGroup(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

// TestIsBusyGroupError verifies BUSYGROUP error detection.
func TestIsBusyGroupError(t *testing.T) {
	if !isBusyGroupError(errors.New("BUSYGROUP something")) {
		t.Error("expected BUSYGROUP error to be detected")
	}

	if isBusyGroupError(errors.New("other error")) {
		t.Error("did not expect BUSYGROUP match")
	}

	if isBusyGroupError(nil) {
		t.Error("did not expect nil error to match")
	}
}
