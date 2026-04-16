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
	"database/sql"
	"fmt"
	"os"

	"github.com/ghdrope/court/internal/archive"
	"github.com/ghdrope/court/internal/prosecutor"
	redisstream "github.com/ghdrope/court/internal/transport/redis"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const defaultDSN = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"
const defaultRedisAddr = "localhost:6379"

// newProsecutorCommand starts the Prosecutor worker.
//
// It consumes stored events and enriches them before forwarding
// to the Court service.
func newProsecutorCommand() *cobra.Command {

	var redisAddr string

	cmd := &cobra.Command{
		Use:   "prosecutor",
		Short: "Start Prosecutor worker",
		Long:  "Prosecutor consumes stored events and enriches Court handled data.",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()
			logger := zap.L().With(zap.String("component", "prosecutor"))

			// Resolve database DSN
			dsn := os.Getenv("DATABASE_URL")
			if dsn == "" {
				dsn = defaultDSN
			}

			// Resolve Redis address
			if redisAddr == "" {
				redisAddr = os.Getenv("REDIS_ADDR")
				if redisAddr == "" {
					redisAddr = defaultRedisAddr
				}
			}

			// Open database connection
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				return fmt.Errorf("db open: %w", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					logger.Error("failed to close db", zap.Error(err))
				}
			}()

			// Ensure DB is reachable
			if err := archive.WaitForDB(ctx, db); err != nil {
				return err
			}

			// Create Redis client
			rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

			client := redisstream.NewProsecutorStreamClient(rdb)
			courtClient := redisstream.NewCourtStreamClient(rdb)

			// Ensure consumer group exists
			if err := client.Client.EnsureGroup(ctx); err != nil {
				return err
			}

			// Initialize prosecutor service
			svc := prosecutor.New(db)
			svc.Publisher = courtClient

			logger.Info("prosecutor worker started")

			// Start consuming events
			return client.ConsumeProsecutor(ctx, svc)
		},
	}

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")

	return cmd
}
