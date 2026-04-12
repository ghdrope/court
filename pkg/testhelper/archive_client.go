package testhelper

import (
	"context"

	archivepb "github.com/ghdrope/court/proto/archive"
	"google.golang.org/grpc"
)

// MockArchiveServiceClient is a test double for the ArchiveService gRPC client.
type MockArchiveServiceClient struct {
	// StoreIncidentFunc is called when StoreIncident is invoked.
	// If nil, the call succeeds by default.
	StoreIncidentFunc func(ctx context.Context, req *archivepb.StoreIncidentRequest) error
}

// StoreIncident simulates the ArchiveService gRPC method.
// It returns a successful response if StoreIncidentFunc returns no error.
func (m *MockArchiveServiceClient) StoreIncident(
	ctx context.Context,
	req *archivepb.StoreIncidentRequest,
	_ ...grpc.CallOption,
) (*archivepb.StoreIncidentResponse, error) {

	if m.StoreIncidentFunc != nil {
		err := m.StoreIncidentFunc(ctx, req)

		return &archivepb.StoreIncidentResponse{
			Success: err == nil,
		}, err
	}

	// Default behavior: successful response
	return &archivepb.StoreIncidentResponse{
		Success: true,
	}, nil
}

// MockArchiveClient is a lightweight mock implementation of the archive repository.
type MockArchiveClient struct {
	// StoreFunc defines the behavior of the Store method.
	// If nil, Store succeeds without error.
	StoreFunc func(ctx context.Context, req *archivepb.StoreIncidentRequest) error
}

// Store simulates persisting an incident into storage.
func (m *MockArchiveClient) Store(ctx context.Context, req *archivepb.StoreIncidentRequest) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(ctx, req)
	}
	return nil
}
