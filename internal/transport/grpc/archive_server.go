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

	"github.com/ghdrope/court/internal/archive"
	"github.com/ghdrope/court/internal/incident"
	pb "github.com/ghdrope/court/proto/incident"
)

// ArchiveServer implements the gRPC ArchiveService.
//
// It acts as the ingestion layer that receives IncidentReports
// over the network and persists them unsing the archive service.
type ArchiveServer struct {
	pb.UnimplementedArchiveServiceServer

	archive *archive.Archive
}

// NewArchiveServer creates a new ArchiveServer instance.
func NewArchiveServer(a *archive.Archive) *ArchiveServer {
	return &ArchiveServer{
		archive: a,
	}
}

// ReceiveStoreIncident handles incoming IncidentReport requests.
func (s *ArchiveServer) ReceiveStoreIncident(
	ctx context.Context,
	req *pb.IncidentReport,
) (*pb.Ack, error) {

	if req == nil {
		return &pb.Ack{Success: false}, nil
	}

	// Convert protobuf -> domain model
	report := convertProtoToDomain(req)

	// Persist incident
	if err := s.archive.StoreIncident(ctx, &report); err != nil {
		return &pb.Ack{Success: false}, err
	}

	return &pb.Ack{Success: true}, nil
}

// convertProtoToDomain maps a protobuf IncidentReport
// into the internal domain representation.
func convertProtoToDomain(req *pb.IncidentReport) incident.IncidentReport {

	events := make([]incident.K8sEvent, len(req.Events))
	for i, e := range req.Events {
		events[i] = incident.K8sEvent{
			Type:    e.Type,
			Reason:  e.Reason,
			Message: e.Message,
		}
	}

	containerIssues := make([]incident.ContainerIssue, len(req.ContainerIssues))
	for i, c := range req.ContainerIssues {
		containerIssues[i] = incident.ContainerIssue{
			Container: c.Container,
			Reason:    c.Reason,
			Logs:      c.Logs,
		}
	}

	return incident.IncidentReport{
		ID:        req.Id,
		Cluster:   req.Cluster,
		Namespace: req.Namespace,
		Pod:       req.Pod,

		Events:          events,
		ContainerIssues: containerIssues,

		ProsecutorCommentary: req.ProsecutorCommentary,
	}
}
