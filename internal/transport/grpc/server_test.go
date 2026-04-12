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
