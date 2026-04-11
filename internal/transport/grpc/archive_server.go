package grpc

import (
	"context"
	"log"

	"github.com/ghdrope/court/internal/archive"
	pb "github.com/ghdrope/court/proto/archive"
)

// ArchiveServer implements ArchiveService gRPC API.
type ArchiveServer struct {
	pb.UnimplementedArchiveServiceServer

	Repo archive.Repository
}

// StoreIncident receives an incident and persists it.
func (s *ArchiveServer) StoreIncident(
	ctx context.Context,
	req *pb.StoreIncidentRequest,
) (*pb.StoreIncidentResponse, error) {

	log.Printf(
		"storing incident event_id=%s pod=%s/%s",
		req.EventId,
		req.Namespace,
		req.PodName,
	)

	if err := s.Repo.Store(ctx, req); err != nil {
		return nil, err
	}

	return &pb.StoreIncidentResponse{
		Status: "stored",
	}, nil
}
