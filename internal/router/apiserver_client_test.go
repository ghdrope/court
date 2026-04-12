package router

import (
	"context"
	"errors"
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	pb "github.com/ghdrope/court/proto/incident"
)

// TestAPIClient_Send_Error ensures gRPC errors are properly propagated.
func TestAPIClient_Send_Error(t *testing.T) {
	mock := &testhelper.MockIncidentServiceClient{
		ReportIncidentFunc: func(ctx context.Context, in *pb.IncidentReport) error {
			return errors.New("grpc failure")
		},
	}

	client := &APIClient{
		client: mock,
	}

	err := client.Send(context.Background(), &pb.IncidentReport{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
