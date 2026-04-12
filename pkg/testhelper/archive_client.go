/*
Copyright 2026 Pedro Cozinheiro.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
