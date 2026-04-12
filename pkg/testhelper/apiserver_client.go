package testhelper

import (
	"context"

	pb "github.com/ghdrope/court/proto/incident"
	"google.golang.org/grpc"
)

// MockAPIClient implements IncidentSender for testing.
type MockAPIClient struct {
	SendFunc func(ctx context.Context, report *pb.IncidentReport) error
}

// Send simulates sending an incident to an external system.
// It returns nil unless SendFunc explicitly returns an error.
func (m *MockAPIClient) Send(ctx context.Context, report *pb.IncidentReport) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, report)
	}
	return nil
}

// MockIncidentServiceClient mocks the gRPC client.
type MockIncidentServiceClient struct {
	ReportIncidentFunc func(ctx context.Context, in *pb.IncidentReport) error
}

// ReportIncident simulates the gRPC method call.
func (m *MockIncidentServiceClient) ReportIncident(
	ctx context.Context,
	in *pb.IncidentReport,
	_ ...grpc.CallOption,
) (*pb.Ack, error) {

	if m.ReportIncidentFunc != nil {
		return &pb.Ack{}, m.ReportIncidentFunc(ctx, in)
	}

	return &pb.Ack{}, nil
}
