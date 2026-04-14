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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/ghdrope/court/internal/archive"
	redisstream "github.com/ghdrope/court/internal/transport/redis"
	"github.com/spf13/cobra"
)

// defaultDSN is used when DATABASE_URL is not provided.
// Enables zero-config
const defaultRedisAddr = "localhost:6379"
const defaultDSN = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"

// newArchiveCommand starts the Archive service.
func newArchiveCommand() *cobra.Command {

	var redisAddr string

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Start gRPC Archive service",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()

			// Resolve Redis address
			if redisAddr == "" {
				redisAddr = os.Getenv("REDIS_ADDR")
				if redisAddr == "" {
					redisAddr = defaultRedisAddr
					zap.L().Warn("REDIS_ADDR not set, using default",
						zap.String("redis-addr", redisAddr))
				}
			}

			// Resolve database DSN
			dsn := os.Getenv("DATABASE_URL")
			if dsn == "" {
				dsn = defaultDSN
				zap.L().Warn("DATABASE_URL not set, using default",
					zap.String("dsn", dsn),
				)
			}

			// Open database connection
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					zap.L().Error("failed to close db", zap.Error(err))
				}
			}()

			// Ensure DB is reachable with retry
			if err := archive.WaitForDB(ctx, db); err != nil {
				return fmt.Errorf("database not ready: %w", err)
			}

			rdb := redis.NewClient(&redis.Options{
				Addr: redisAddr,
			})

			baseClient := redisstream.NewClient(rdb)

			// inbound stream (Officer -> Archive)
			incidentClient := redisstream.NewIncidentStreamClient(baseClient)

			// outbound stream (Archive -> Prosecutor)
			prosecutorClient := redisstream.NewProsecutorStreamClient(baseClient)

			// Archive now uses generic publisher
			arch := archive.New(db, prosecutorClient)

			// Ensure schema exists
			if err := arch.InitSchema(ctx); err != nil {
				return fmt.Errorf("init schema: %w", err)
			}

			if err := incidentClient.EnsureGroup(ctx); err != nil {
				return err
			}

			zap.L().Info("archive worker started")

			return incidentClient.ConsumeLoop(ctx, arch)
		},
	}

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")

	return cmd
}
