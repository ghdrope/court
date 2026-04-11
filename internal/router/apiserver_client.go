package router

import (
	"context"

	pb "github.com/ghdrope/court/proto/incident"
	"google.golang.org/grpc"
)

// APIClient sends incidents to API server.
type APIClient struct {
	client pb.IncidentServiceClient
}

// NewAPIClient creates a new gRPC client.
func NewAPIClient(conn *grpc.ClientConn) *APIClient {
	return &APIClient{
		client: pb.NewIncidentServiceClient(conn),
	}
}

// Send sends an incident report to API server.
func (c *APIClient) Send(ctx context.Context, report *pb.IncidentReport) error {
	_, err := c.client.ReportIncident(ctx, report)
	return err
}
