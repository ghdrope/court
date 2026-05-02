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
	"fmt"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

// redisClient defines the minimal Redis operations required by StreamClient.
//
// It allows decoupling from the concrete *goredis.Client type,
// enabling easier testing via mocks or fakes.
type redisClient interface {
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) *goredis.StatusCmd
	XReadGroup(ctx context.Context, args *goredis.XReadGroupArgs) *goredis.XStreamSliceCmd
	XAck(ctx context.Context, stream, group string, ids ...string) *goredis.IntCmd
}

// StreamClient provides access to Redis Streams operations.
//
// It is responsible for stream-level operations such as
// consumer group creation and message consumption.
type StreamClient struct {
	rdb redisClient
	cfg Config
}

// NewStreamClient creates a new StreamClient instance.
func NewStreamClient(rdb *goredis.Client, cfg Config) *StreamClient {
	return &StreamClient{
		rdb: rdb,
		cfg: cfg,
	}
}

// EnsureGroup creates the consumer group if it does not exist.
//
// It is idempotent and safe to call multiple times during startup.
func (c *StreamClient) EnsureGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(
		ctx,
		c.cfg.Stream,
		c.cfg.Group,
		"0",
	).Err()

	if err != nil && !isBusyGroupError(err) {
		return fmt.Errorf("redis: create consumer group: %w", err)
	}

	return nil
}

// isBusyGroupError checks whether Redis returned a BUSYGROUP error.
//
// This happens when the consumer group already exists.
func isBusyGroupError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "BUSYGROUP")
}
