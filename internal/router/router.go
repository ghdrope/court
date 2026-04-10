package router

import (
	"context"
	"log"

	pb "github.com/ghdrope/court/proto/incident"
)

// Router is a simple routing implementation.
type Router struct {
	CourtClient CourtClient
}

// CourtClient defines downstream Court communication.
type CourtClient interface {
	Send(ctx context.Context, report *pb.IncidentReport) error
}

// Route forwards the incident to the appropriate service.
func (r *Router) Route(ctx context.Context, report *pb.IncidentReport) error {

	log.Printf("routing incident to court: %s", report.EventId)

	return r.CourtClient.Send(ctx, report)
}
