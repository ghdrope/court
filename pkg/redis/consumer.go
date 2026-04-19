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
// The handler receives the raw JSON payload extracted from the Redis Stream.
// If an error is returned, the message will NOT be acknowledged.
type HandlerFunc func(ctx context.Context, data []byte) error

// Consume starts a blocking loop that continuously reads messages
// from the Redis Stream consumer group and dispatches them to the handler.
//
// Messages are acknowledged only after successful processing.
//
// This function is blocking and exits when the context is cancelled.
func (c *StreamClient) Consume(ctx context.Context, handler HandlerFunc) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		res, err := c.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
			Group:    c.cfg.Group,
			Consumer: c.cfg.Consumer,
			Streams:  []string{c.cfg.Stream, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}

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

				if err := c.rdb.XAck(ctx, c.cfg.Stream, c.cfg.Group, msg.ID).Err(); err != nil {
					continue
				}
			}
		}
	}
}
