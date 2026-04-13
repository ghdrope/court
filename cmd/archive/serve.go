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
	"net"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/ghdrope/court/internal/archive"
	grpcserver "github.com/ghdrope/court/internal/transport/grpc"
	incidentpb "github.com/ghdrope/court/proto/incident"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// defaultDSN is used when DATABASE_URL is not provided.
// Enables zero-config
const defaultDSN = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"

// newServeCommand starts the Archive gRPC server.
func newServeCommand() *cobra.Command {

	var port string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC Archive server",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()

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

			// Initialize archive service
			arch := archive.New(db)

			// Ensure schema exists
			if err := arch.InitSchema(ctx); err != nil {
				return fmt.Errorf("init schema: %w", err)
			}

			// Start TCP listener
			lis, err := net.Listen("tcp", ":"+port)
			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			// Create gRPC server
			grpcServer := grpc.NewServer()

			// Register ArchiveService
			incidentpb.RegisterArchiveServiceServer(
				grpcServer,
				grpcserver.NewArchiveServer(arch),
			)

			zap.L().Info("archive service running",
				zap.String("port", port))

			// Start server in background
			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					zap.L().Error("grpc server error", zap.Error(err))
				}
			}()

			// Wait for shutdown signal
			<-ctx.Done()

			zap.L().Info("shutting down archive server")

			grpcServer.GracefulStop()

			return nil
		},
	}

	cmd.Flags().StringVar(&port, "port", "50052", "gRPC port")

	return cmd
}
