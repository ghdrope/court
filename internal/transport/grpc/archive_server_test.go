package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	pb "github.com/ghdrope/court/proto/archive"
)

// TestArchiveServer_StoreIncident tests the StoreIncident gRPC handler.
func TestArchiveServer_StoreIncident(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.StoreIncidentRequest
		repoErr   error
		expectErr bool
		expectOK  bool
	}{
		{
			name: "successfully stores incident",
			req: &pb.StoreIncidentRequest{
				Id:        "incident-1",
				Namespace: "default",
				PodName:   "pod-1",
			},
			repoErr:   nil,
			expectErr: false,
			expectOK:  true,
		},
		{
			name: "repository returns error",
			req: &pb.StoreIncidentRequest{
				Id:        "incident-2",
				Namespace: "kube-system",
				PodName:   "pod-2",
			},
			repoErr:   errors.New("db failure"),
			expectErr: true,
			expectOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &testhelper.MockArchiveClient{
				StoreFunc: func(ctx context.Context, req *pb.StoreIncidentRequest) error {
					if req.Id == "" {
						t.Errorf("expected request id to be set")
					}
					return tt.repoErr
				},
			}

			s := &ArchiveServer{
				Repo: mockRepo,
			}

			resp, err := s.StoreIncident(context.Background(), tt.req)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if resp != nil {
					t.Fatalf("expected nil response on error, got %v", resp)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp == nil {
					t.Fatalf("expected response but got nil")
				}
			}
		})
	}
}
