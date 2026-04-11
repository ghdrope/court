package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ghdrope/court/internal/archive"
	grpcserver "github.com/ghdrope/court/internal/transport/grpc"
	archivepb "github.com/ghdrope/court/proto/archive"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// newArchiveCommand starts the archive.
func newServeCommand() *cobra.Command {

	var port string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC Archive server",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()

			lis, err := net.Listen("tcp", ":"+port)
			if err != nil {
				log.Fatal(err)
			}

			dsn := os.Getenv("DATABASE_URL")
			if dsn == "" {
				return fmt.Errorf("DATABASE_URL is not set")
			}

			db, err := sql.Open("pgx", dsn)
			if err != nil {
				return err
			}
			defer func() {
				_ = db.Close()
			}()

			if err := db.Ping(); err != nil {
				return err
			}

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
