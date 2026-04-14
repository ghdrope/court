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

package redisstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ghdrope/court/internal/archive"
	"github.com/redis/go-redis/v9"
)

// Prosecutor stream configuration.
const (
	ProsecutorStream = "prosecutor:stream"
)

// ProsecutorStreamClient publishes stored events
// to Prosecutor service.
type ProsecutorStreamClient struct {
	*Client
}

// NewProsecutorStreamClient creates a new instance.
func NewProsecutorStreamClient(base *Client) *ProsecutorStreamClient {
	return &ProsecutorStreamClient{Client: base}
}

// PublishStored publishes a generic stored event.
func (c *ProsecutorStreamClient) PublishStored(
	ctx context.Context,
	event archive.StoredEvent,
) error {

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal stored event: %w", err)
	}

	if err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: ProsecutorStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err(); err != nil {
		return fmt.Errorf("xadd prosecutor event: %w", err)
	}

	return nil
}
