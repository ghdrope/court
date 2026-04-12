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
	"log"
	"net"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/ghdrope/court/internal/archive"
	grpcserver "github.com/ghdrope/court/internal/transport/grpc"
	archivepb "github.com/ghdrope/court/proto/archive"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// newServeCommand starts the archive.
func newServeCommand() *cobra.Command {

	var port string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC Archive server",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()

			lis, err := net.Listen("tcp", ":"+port)
			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			dsn := os.Getenv("DATABASE_URL")
			if dsn == "" {
				return fmt.Errorf("DATABASE_URL is not set")
			}

			db, err := sql.Open("pgx", dsn)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					zap.L().Error("failed to close db", zap.Error(err))
				}
			}()

			if err := db.Ping(); err != nil {
				return fmt.Errorf("ping database: %w", err)
			}

			if err := archive.InitSchema(ctx, db); err != nil {
				return fmt.Errorf("init schema: %w", err)
			}

			repo := archive.NewPostgresRepository(db)

			grpcServer := grpc.NewServer()

			archivepb.RegisterArchiveServiceServer(grpcServer, &grpcserver.ArchiveServer{
				Repo: repo,
			})

			zap.L().Info("archive service running", zap.String("port", port))

			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					zap.L().Error("grpc server error", zap.Error(err))
				}
			}()

			<-ctx.Done()

			log.Println("shutting down Archive server")
			grpcServer.GracefulStop()

			return nil
		},
	}

	cmd.Flags().StringVar(&port, "port", "50052", "gRPC port")

	return cmd
}
