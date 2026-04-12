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
