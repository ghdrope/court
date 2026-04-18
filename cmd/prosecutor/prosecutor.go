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

package main

import (
	"fmt"
	"os"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/prosecutor"
	redisstream "github.com/ghdrope/court/internal/transport/redis"
	"github.com/ghdrope/court/pkg/env"
	"github.com/ghdrope/court/pkg/postgres"
	redispkg "github.com/ghdrope/court/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultDSN       = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"
	defaultRedisAddr = "localhost:6379"
)

// newProsecutorCommand initializes the Prosecutor worker process.
//
// The Prosecutor is responsible for:
//   - Consuming "incident.created" events from Redis Streams
//   - Loading full IncidentReports from PostgreSQL
//   - Performing post-processing analysis
//   - Persisting analysis results back into the archive
//   - Publishing downstream events for further processing
func newProsecutorCommand() *cobra.Command {

	var (
		redisAddr string
		dsn       string
	)

	cmd := &cobra.Command{
		Use:   "prosecutor",
		Short: "Start Prosecutor worker",
		Long:  "Prosecutor consumes incident events and enriches them with analysis before persistence.",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()
			logger := zap.L().With(zap.String("component", "prosecutor"))

			// Resolve configuration (flag > env > default)
			databaseURL := env.FirstNonEmpty(
				dsn,
				env.Get("DATABASE_URL", defaultDSN),
			)

			redisAddress := env.FirstNonEmpty(
				redisAddr,
				env.Get("REDIS_ADDR", defaultRedisAddr),
			)

			// --- PostgreSQL ---
			// Initialize PostgreSQL connection.
			// This is required for persisting incident storage.
			db, err := postgres.Open(databaseURL)
			if err != nil {
				return fmt.Errorf("db open: %w", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					logger.Error("failed to close database", zap.Error(err))
				}
			}()

			// Ensure DB is ready before processing events
			if err := postgres.PingWithRetry(cmd.Context(), db); err != nil {
				return fmt.Errorf("db not ready: %w", err)
			}

			// Repository layer.
			repo := incident.NewRepository(db)

			// --- Redis ---
			rdb := goredis.NewClient(&goredis.Options{
				Addr: redisAddress,
			})

			hostname, _ := os.Hostname()
			consumerName := fmt.Sprintf("prosecutor-%s", hostname)

			streamClient := redispkg.NewStreamClient(rdb, redispkg.Config{
				Stream:   "incident.created",
				Group:    "prosecutor-group",
				Consumer: consumerName,
			})

			svc := prosecutor.New(repo, logger)

			consumer := redisstream.NewIncidentCreatedConsumer(streamClient, logger)

			logger.Info("prosecutor started",
				zap.String("consumer", consumerName),
			)

			return consumer.Start(ctx, svc)
		},
	}

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")
	cmd.Flags().StringVar(&dsn, "database-url", "", "PostgreSQL DSN")

	return cmd
}
