package testhelper

import (
	"context"

	pb "github.com/ghdrope/court/proto/incident"
)

// MockRouterAdapter adapts a function into the Router interface.
type MockRouterAdapter struct {
	Fn func(ctx context.Context, report *pb.IncidentReport) error
}

// Route implements Router.
func (m *MockRouterAdapter) Route(ctx context.Context, report *pb.IncidentReport) error {
	return m.Fn(ctx, report)
}
