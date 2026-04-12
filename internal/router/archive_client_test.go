package router

import (
	"context"
	"errors"
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	archivepb "github.com/ghdrope/court/proto/archive"
)

// TestGRPCArchiveClient_Store_Success ensures successful request to archive service.
func TestGRPCArchiveClient_Store_Success(t *testing.T) {
	mock := &testhelper.MockArchiveServiceClient{
		StoreIncidentFunc: func(ctx context.Context, req *archivepb.StoreIncidentRequest) error {

			if req.Id != "id-1" {
				t.Errorf("unexpected id: %s", req.Id)
			}

			return nil
		},
	}

	client := &GRPCArchiveClient{
		client: mock,
	}

	err := client.Store(context.Background(), &archivepb.StoreIncidentRequest{
		Id: "id-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGRPCArchiveClient_Store_Error ensures error propagation.
func TestGRPCArchiveClient_Store_Error(t *testing.T) {
	mock := &testhelper.MockArchiveServiceClient{
		StoreIncidentFunc: func(ctx context.Context, req *archivepb.StoreIncidentRequest) error {
			return errors.New("archive failure")
		},
	}

	client := &GRPCArchiveClient{
		client: mock,
	}

	err := client.Store(context.Background(), &archivepb.StoreIncidentRequest{
		Id: "id-1",
	})

	if err == nil {
		t.Fatal("expected error")
	}
}
