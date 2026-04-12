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
	"log"
	"net"

	"github.com/ghdrope/court/internal/router"
	grpcserver "github.com/ghdrope/court/internal/transport/grpc"
	pb "github.com/ghdrope/court/proto/incident"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// newServeCommand starts the stateless gRPC API server.
func newServeCommand() *cobra.Command {

	var port string
	var archiveAddr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC API server",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()

			lis, err := net.Listen("tcp", ":"+port)
			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			// connect to Archive service
			conn, err := grpc.NewClient(
				archiveAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return fmt.Errorf("connect archive: %w", err)
			}
			defer func() {
				if err := conn.Close(); err != nil {
					zap.L().Error("failed to close gRPC connection", zap.Error(err))
				}
			}()

			archiveClient := router.NewGRPCArchiveClient(conn)

			r := &router.Router{
				ArchiveClient: archiveClient,
			}

			s := &grpcserver.Server{
				Router: r,
			}

			grpcServer := grpc.NewServer()

			pb.RegisterIncidentServiceServer(grpcServer, s)

			zap.L().Info("API server listening", zap.String("port", port))

			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					zap.L().Error("grpc server error", zap.Error(err))
				}
			}()

			<-ctx.Done()

			log.Println("shutting down API server")
			grpcServer.GracefulStop()

			return nil
		},
	}

	cmd.Flags().StringVar(&port, "port", "50051", "gRPC port")
	cmd.Flags().StringVar(&archiveAddr, "archive-addr", "localhost:50052", "Archive gRPC address")

	return cmd
}
