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

	"github.com/ghdrope/court/internal/court"
	"github.com/ghdrope/court/internal/suit"
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
	defaultRedisAddr = "localhost:6379"
	defaultDSN       = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"
)

// newCourtCommand initializes the Court worker process.
//
// The Court service is responsible for:
//   - Consuming "incident.analyzed" events from Redis Streams
//   - Creating Suit records based on analyzed incidents
//   - Ensuring idempotent creation of suits per incident
//   - Persisting legal case state into PostgreSQL
func newCourtCommand() *cobra.Command {

	var (
		redisAddr string
		dsn       string
	)

	cmd := &cobra.Command{
		Use:   "court",
		Short: "Start Court worker",
		Long:  "Court consumes analyzed incidents and creates suits in the archive system.",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()
			logger := zap.L().With(zap.String("component", "court"))

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
			// This is required for persisting suit storage.
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
			repo := suit.NewRepository(db)

			// Ensure schema is up to date
			if err := repo.InitSchema(cmd.Context()); err != nil {
				return err
			}

			// --- Redis ---
			// Initialize Redis client used for consuming incident.analyzed stream
			rdb := goredis.NewClient(&goredis.Options{
				Addr: redisAddress,
			})

			hostname, _ := os.Hostname()
			consumerName := fmt.Sprintf("court-%s", hostname)

			streamClient := redispkg.NewStreamClient(rdb, redispkg.Config{
				Stream:   "incident.analyzed",
				Group:    "court-group",
				Consumer: consumerName,
			})

			// Court service handles Suit lifecycle creation
			svc := court.New(repo, logger)

			// Redis Stream consumer for analyzed incidents
			consumer := redisstream.NewIncidentAnalyzedConsumer(streamClient, logger)

			logger.Info("court started",
				zap.String("consumer", consumerName),
			)

			return consumer.Start(ctx, svc)
		},
	}

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")
	cmd.Flags().StringVar(&dsn, "database-url", "", "PostgreSQL DSN")

	return cmd
}
