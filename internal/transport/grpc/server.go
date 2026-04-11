package grpc

import (
	"context"
	"log"

	pb "github.com/ghdrope/court/proto/incident"
)

// Server implements the gRPC API server.
type Server struct {
	pb.UnimplementedIncidentServiceServer

	Router Router
}

// Router defines where to forward incidents.
type Router interface {
	Route(ctx context.Context, report *pb.IncidentReport) error
}

// ReportIncident receives an incident and forwards it.
func (s *Server) ReportIncident(ctx context.Context, report *pb.IncidentReport) (*pb.Ack, error) {

	log.Printf("received incident: %s/%s id=%s",
		report.Namespace,
		report.PodName,
		report.Id,
	)

	if err := s.Router.Route(ctx, report); err != nil {
		return nil, err
	}

	return &pb.Ack{Success: true}, nil
}
