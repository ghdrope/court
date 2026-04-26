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
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// HandlerFunc defines the function signature for processing stream messages.
//
// The handler receives the raw payload extracted from the Redis Stream.
// If an error is returned, the message will not be acknowledged.
type HandlerFunc func(ctx context.Context, data []byte) error

// Consume starts a blocking loop that continuously reads messages
// from the Redis Stream consumer group and dispatches them to the handler.
//
// Messages are acknowledged only after successful processing.
//
// This function blocks until the provided context is cancelled.
func (c *StreamClient) Consume(ctx context.Context, handler HandlerFunc) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := c.consumeBatch(ctx, handler); err != nil {
			// Respect context cancellation explicitly
			if errors.Is(err, context.Canceled) {
				return err
			}

			// Small backoff to avoid tight retry loops on transient errors
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// consumeBatch performs a single XREADGROUP operation and processes messages.
func (c *StreamClient) consumeBatch(ctx context.Context, handler HandlerFunc) error {
	res, err := c.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    c.cfg.Group,
		Consumer: c.cfg.Consumer,
		Streams:  []string{c.cfg.Stream, ">"},
		Count:    int64(c.cfg.BatchSize),
		Block:    c.cfg.BlockTime,
	}).Result()

	if err != nil {
		return err
	}

	for _, stream := range res {
		for _, msg := range stream.Messages {
			c.processMessage(ctx, handler, msg)
		}
	}

	return nil
}

// processMessage handles a single Redis Stream message.
//
// It extracts the payload, executes the handler, and acknowledges
// the message only if processing succeeds.
func (c *StreamClient) processMessage(
	ctx context.Context,
	handler HandlerFunc,
	msg goredis.XMessage,
) {
	raw, ok := msg.Values["payload"].(string)
	if !ok {
		return
	}

	// Execute handler
	if err := handler(ctx, []byte(raw)); err != nil {
		return
	}

	// Acknowledge message after successful processing
	_ = c.rdb.XAck(ctx, c.cfg.Stream, c.cfg.Group, msg.ID).Err()
}
