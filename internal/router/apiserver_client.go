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

package router

import (
	"context"

	pb "github.com/ghdrope/court/proto/incident"
	"google.golang.org/grpc"
)

// IncidentSender defines how incidents are sent to an external system.
type IncidentSender interface {
	Send(ctx context.Context, report *pb.IncidentReport) error
}

// APIClient sends incidents to API server.
type APIClient struct {
	client pb.IncidentServiceClient
}

// NewAPIClient creates a new gRPC client.
func NewAPIClient(conn *grpc.ClientConn) *APIClient {
	return &APIClient{
		client: pb.NewIncidentServiceClient(conn),
	}
}

// Send sends an incident report to API server.
func (c *APIClient) Send(ctx context.Context, report *pb.IncidentReport) error {
	_, err := c.client.ReportIncident(ctx, report)
	return err
}
