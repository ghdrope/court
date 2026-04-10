package router

import (
	"context"
	"log"

	pb "github.com/ghdrope/court/proto/incident"
)

// MockCourtClient is a temporary implementation.
type MockCourtClient struct{}

func (m *MockCourtClient) Send(ctx context.Context, report *pb.IncidentReport) error {
	log.Printf("court received incident: %s/%s", report.Namespace, report.PodName)
	return nil
}
