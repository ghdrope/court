package router

import (
	"context"

	archivepb "github.com/ghdrope/court/proto/archive"
	"google.golang.org/grpc"
)

// GRPCArchiveClient implements ArchiveClient over gRPC.
type GRPCArchiveClient struct {
	client archivepb.ArchiveServiceClient
}

// NewGRPCArchiveClient creates a new Archive gRPC client.
func NewGRPCArchiveClient(conn *grpc.ClientConn) *GRPCArchiveClient {
	return &GRPCArchiveClient{
		client: archivepb.NewArchiveServiceClient(conn),
	}
}

// Store sends incident to Archive service.
func (c *GRPCArchiveClient) Store(
	ctx context.Context,
	req *archivepb.StoreIncidentRequest,
) error {

	_, err := c.client.StoreIncident(ctx, req)
	return err
}
