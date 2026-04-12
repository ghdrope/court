package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	pb "github.com/ghdrope/court/proto/incident"
)

// TestServer_ReportIncident tests the ReportIncident gRPC handler.
func TestServer_ReportIncident(t *testing.T) {
	tests := []struct {
		name      string
		report    *pb.IncidentReport
		routerErr error
		expectErr bool
	}{
		{
			name: "successfully routes incident",
			report: &pb.IncidentReport{
				Id:        "incident-1",
				Namespace: "default",
				PodName:   "pod-1",
			},
			routerErr: nil,
			expectErr: false,
		},
		{
			name: "router returns error",
			report: &pb.IncidentReport{
				Id:        "incident-2",
				Namespace: "default",
				PodName:   "pod-2",
			},
			routerErr: errors.New("routing failed"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &testhelper.MockRouterAdapter{
				Fn: func(ctx context.Context, r *pb.IncidentReport) error {
					return tt.routerErr
				},
			}

			s := &Server{
				Router: router,
			}

			resp, err := s.ReportIncident(context.Background(), tt.report)

			if tt.expectErr && err == nil {
				t.Fatalf("expected error but got nil")
			}

			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.expectErr && resp == nil {
				t.Fatalf("expected response but got nil")
			}

			if resp != nil && resp.Success != !tt.expectErr {
				t.Errorf("unexpected success value: %v", resp.Success)
			}
		})
	}
}
