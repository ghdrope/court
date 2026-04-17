package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

// StreamClient provides operations over Redis Streams,
// including publishing, consumer group management, and consumption.
type StreamClient struct {
	rdb *goredis.Client
	cfg Config
}

// NewStreamClient creates a new StreamClient instance.
func NewStreamClient(rdb *goredis.Client, cfg Config) *StreamClient {
	return &StreamClient{
		rdb: rdb,
		cfg: cfg,
	}
}

// Publish sends a payload into the configured Redis Stream.
// The payload is serialized as JSON and stored under the "payload" field.
func (c *StreamClient) Publish(ctx context.Context, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return c.rdb.XAdd(ctx, &goredis.XAddArgs{
		Stream: c.cfg.Stream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err()
}

// EnsureGroup creates the consumer group if it does not exist.
// Operation is idempotent.
//
// It uses XGroupCreateMkStream to also create the stream if missing.
func (c *StreamClient) EnsureGroup(ctx context.Context) error {
	err := c.rdb.XGroupCreateMkStream(
		ctx,
		c.cfg.Stream,
		c.cfg.Group,
		"0",
	).Err()

	if err != nil && err != goredis.Nil && !isBusyGroupError(err) {
		return fmt.Errorf("create consumer group: %w", err)
	}

	return nil
}

// isBusyGroupError checks whether Redis returned a BUSYGROUP error.
// This happens when the consumer group already exists.
func isBusyGroupError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "BUSYGROUP")
}
