package redis

import (
	"context"
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
