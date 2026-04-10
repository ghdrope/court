package main

import (
	"log"
	"net"

	"github.com/ghdrope/court/internal/router"
	grpcserver "github.com/ghdrope/court/internal/transport/grpc"
	pb "github.com/ghdrope/court/proto/incident"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// newServeCommand starts the stateless gRPC API server.
func newServeCommand() *cobra.Command {

	var port string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC API server",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()

			lis, err := net.Listen("tcp", ":"+port)
			if err != nil {
				return err
			}

			grpcServer := grpc.NewServer()

			r := &router.Router{
				CourtClient: &router.MockCourtClient{},
			}

			s := &grpcserver.Server{
				Router: r,
			}

			pb.RegisterIncidentServiceServer(grpcServer, s)

			log.Printf("API server listening on :%s", port)

			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					log.Printf("grpc error: %v", err)
				}
			}()

			<-ctx.Done()

			log.Println("shutting down API server...")
			grpcServer.GracefulStop()

			return nil
		},
	}

	cmd.Flags().StringVar(&port, "port", "50051", "gRPC port")

	return cmd
}
