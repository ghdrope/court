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
	"fmt"
	"time"

	"github.com/ghdrope/court/internal/incident"
	incidentpb "github.com/ghdrope/court/proto/incident"
)

// ArchiveClient defines the contract used by the Officer to send incidents.
type ArchiveClient interface {
	Send(ctx context.Context, report *incident.IncidentReport) error
}

// archiveClient is a gRPC implementation of ArchiveClient.
type archiveClient struct {
	client incidentpb.ArchiveServiceClient
}

// NewArchiveClient creates a new ArchiveClient backed by gRPC.
func NewArchiveClient(client incidentpb.ArchiveServiceClient) ArchiveClient {
	return &archiveClient{
		client: client,
	}
}

// Send transmits an IncidentReport to the Archive service.
func (c *archiveClient) Send(ctx context.Context, report *incident.IncidentReport) error {

	if report == nil {
		return fmt.Errorf("report is nil")
	}

	// Map domain -> protobuf
	pbReport := &incidentpb.IncidentReport{
		Id:        report.ID,
		Cluster:   report.Cluster,
		Namespace: report.Namespace,
		Pod:       report.Pod,
	}

	for _, e := range report.Events {
		pbReport.Events = append(pbReport.Events, &incidentpb.K8SEvent{
			Type:    e.Type,
			Reason:  e.Reason,
			Message: e.Message,
		})
	}

	for _, ci := range report.ContainerIssues {
		pbReport.ContainerIssues = append(pbReport.ContainerIssues, &incidentpb.ContainerIssue{
			Container: ci.Container,
			Reason:    ci.Reason,
			Logs:      ci.Logs,
		})
	}

	pbReport.ProsecutorCommentary = report.ProsecutorCommentary

	// Timeout per request
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Call gRPC
	resp, err := c.client.ReceiveStoreIncident(ctx, pbReport)
	if err != nil {
		return fmt.Errorf("send incident: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("archive rejected incident")
	}

	return nil
}
