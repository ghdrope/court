package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/ghdrope/court/internal/archive"
	grpcserver "github.com/ghdrope/court/internal/transport/grpc"
	archivepb "github.com/ghdrope/court/proto/archive"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// newArchiveCommand starts the archive.
func newArchiveCommand() *cobra.Command {

	var port string

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Start gRPC Archive server",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()

			lis, err := net.Listen("tcp", ":"+port)
			if err != nil {
				log.Fatal(err)
			}

			db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				_ = db.Close()
			}()

			repo := archive.NewPostgresRepository(db)

			grpcServer := grpc.NewServer()

			archivepb.RegisterArchiveServiceServer(grpcServer, &grpcserver.ArchiveServer{
				Repo: repo,
			})

			log.Printf("archive service running on :%s", port)

			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					log.Printf("grpc error: %v", err)
				}
			}()

			<-ctx.Done()

			log.Println("shutting down Archive server...")
			grpcServer.GracefulStop()

			return err
		},
	}

	cmd.Flags().StringVar(&port, "port", "50052", "gRPC port")

	return cmd
}
