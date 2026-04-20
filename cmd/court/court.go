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
	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/suit"

	redisstream "github.com/ghdrope/court/internal/transport/redis"

	"github.com/ghdrope/court/pkg/env"
	"github.com/ghdrope/court/pkg/github"
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
//   - Creating Suit records for validated incidents
//   - Creating GitHub issues from incident metadata
func newCourtCommand() *cobra.Command {

	var (
		redisAddr   string
		dsn         string
		githubToken string
	)

	cmd := &cobra.Command{
		Use:   "court",
		Short: "Start Court worker",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()
			logger := zap.L().With(zap.String("component", "court"))

			// --- Configuration resolution ---
			// Priority: CLI flags > environment variables > default
			databaseURL := env.FirstNonEmpty(dsn, env.Get("DATABASE_URL", defaultDSN))

			redisAddress := env.FirstNonEmpty(redisAddr, env.Get("REDIS_ADDR", defaultRedisAddr))

			ghToken := env.FirstNonEmpty(githubToken, env.Get("GITHUB_TOKEN", ""))

			// --- PostgreSQL initialization ---
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

			// Ensure database is reachable before starting consumers.
			if err := postgres.PingWithRetry(cmd.Context(), db); err != nil {
				return fmt.Errorf("db not ready: %w", err)
			}

			// --- Repositories ---
			// Repository handles persistence of suits.
			suitRepo := suit.NewRepository(db)
			incidentRepo := incident.NewRepository(db)

			// Ensure schema is applied before processing events.
			if err := suitRepo.InitSchema(ctx); err != nil {
				return err
			}

			// --- Redis setup ---
			// Initialize Redis client used for consuming incident.analyzed stream
			rdb := goredis.NewClient(&goredis.Options{
				Addr: redisAddress,
			})

			hostname, _ := os.Hostname()
			consumerName := fmt.Sprintf("court-%s", hostname)

			incidentCreatedClient := redispkg.NewStreamClient(rdb, redispkg.Config{
				Stream:   redisstream.IncidentCreatedStream,
				Group:    redisstream.CourtGroup,
				Consumer: consumerName,
			})

			// --- GitHub client ---
			// Optional integration used to publish incidents as GitHub issues.
			var ghClient court.GitHubClient

			if ghToken != "" {
				logger.Info("github integration enabled")
				ghClient = github.NewClient(ghToken)
			}

			// --- Court service ---
			// Court service orchestrates suit creation and external side-effects.
			svc := court.New(suitRepo, ghClient, logger)

			// --- Consumer ---
			// Consumer processes created incidents and triggers suit creation.
			consumer := redisstream.NewIncidentCreatedConsumer(incidentCreatedClient, incidentRepo, logger)

			logger.Info("court started")

			// Start event consumption loop.
			return consumer.Start(ctx, svc)
		},
	}

	// --- CLI flags ---
	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")
	cmd.Flags().StringVar(&dsn, "database-url", "", "PostgreSQL DSN")
	cmd.Flags().StringVar(&githubToken, "github-token", "", "GitHub API token")

	return cmd
}
