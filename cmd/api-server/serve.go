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
