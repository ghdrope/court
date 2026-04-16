package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// HandlerFunc defines the function signature for processing stream messages.
//
// The handler receives the raw JSON payload extracted from the Redis Stream.
// If an error is returned, the message will NOT be acknowledged.
type HandlerFunc func(ctx context.Context, data []byte) error

// Consume starts a blocking loop that continuously reads messages
// from the Redis Stream consumer group and dispatches them to the handler.
//
// Messages are acknowledged only after successful processing.
//
// This function is blocking and should typically run in a goroutine.
// It exits when the provided context is cancelled.
func (c *StreamClient) Consume(ctx context.Context, handler HandlerFunc) error {

	for {
		res, err := c.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
			Group:    c.cfg.Group,
			Consumer: c.cfg.Consumer,
			Streams:  []string{c.cfg.Stream, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			// Small backoff to avoid tight retry loops
			time.Sleep(500 * time.Millisecond)
			continue
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {

				raw, ok := msg.Values["payload"].(string)
				if !ok {
					continue
				}

				if err := handler(ctx, []byte(raw)); err != nil {
					continue
				}

				_ = c.rdb.XAck(ctx, c.cfg.Stream, c.cfg.Group, msg.ID).Err()
			}
		}
	}
}
