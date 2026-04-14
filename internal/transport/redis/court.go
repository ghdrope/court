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

// CourtStream is the output stream for downstream Case creation.
const (
	CourtStream = "court:stream"
)

// CourtStreamClient publishes prosecutor completion events.
type CourtStreamClient struct {
	*Client
}

// NewCourtStreamClient creates a new court stream publisher.
func NewCourtStreamClient(base *Client) *CourtStreamClient {
	return &CourtStreamClient{Client: base}
}

// PublishedStored emits a prosecutor.finished event.
func (c *CourtStreamClient) PublishStored(
	ctx context.Context,
	event archive.StoredEvent,
) error {

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal court event: %w", err)
	}

	if err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: CourtStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err(); err != nil {
		return fmt.Errorf("xadd court event: %w", err)
	}

	return nil
}
