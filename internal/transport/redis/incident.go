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

	"github.com/ghdrope/court/internal/archive"
	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Stream configuration for incident ingestion.
const (
	IncidentStream   = "incident:stream"
	IncidentGroup    = "incident-archive-group"
	IncidentConsumer = "archive-1"
)

// IncidentStreamClient handles publishing and consuming incident events.
type IncidentStreamClient struct {
	Client *redis.StreamClient
}

// NewIncidentStreamClient creates a new IncidentStreamClient.
func NewIncidentStreamClient(rdb *goredis.Client) *IncidentStreamClient {
	return &IncidentStreamClient{
		Client: redis.NewStreamClient(rdb, redis.Config{
			Stream:   IncidentStream,
			Group:    IncidentGroup,
			Consumer: IncidentConsumer,
		}),
	}
}

// SendIncident implements the officer.ArchiveClient interface.
//
// It forwards the incident report to the underlying Redis stream.
func (c *IncidentStreamClient) Send(ctx context.Context, r *incident.IncidentReport) error {
	return c.PublishIncident(ctx, r)
}

// PublishIncident publishes an IncidentReport into the Stream.
func (c *IncidentStreamClient) PublishIncident(ctx context.Context, r *incident.IncidentReport) error {
	return c.Client.Publish(ctx, r)
}

// ConsumeIncident reads incident events and persists them into the archive.
//
// Each message is expected to be a serialized IncidentReport.
// Messages are acknowledged only after successful persistence.
func (c *IncidentStreamClient) ConsumeIncident(ctx context.Context, arch *archive.Archive) error {

	logger := zap.L().With(zap.String("component", "incident-consumer"))

	return c.Client.Consume(ctx, func(ctx context.Context, data []byte) error {

		var r incident.IncidentReport
		if err := json.Unmarshal(data, &r); err != nil {
			logger.Error("invalid payload", zap.Error(err))
			return err
		}

		if err := arch.StoreIncident(ctx, &r); err != nil {
			logger.Error("failed to persist incident", zap.Error(err))
			return err
		}

		return nil
	})
}
