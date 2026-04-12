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
	"fmt"
	"log"

	archivepb "github.com/ghdrope/court/proto/archive"
	pb "github.com/ghdrope/court/proto/incident"
)

// Router forwards incidents to Archive service.
type Router struct {
	ArchiveClient ArchiveClient
}

// ArchiveClient defines communication with Archive service.
type ArchiveClient interface {
	Store(ctx context.Context, req *archivepb.StoreIncidentRequest) error
}

// Route converts and forwards the incident to Archive.
func (r *Router) Route(ctx context.Context, report *pb.IncidentReport) error {

	if report == nil {
		return fmt.Errorf("incident report is nil")
	}

	log.Printf("routing incident id=%s pod=%s/%s",
		report.Id,
		report.Namespace,
		report.PodName,
	)

	req := &archivepb.StoreIncidentRequest{
		Id:        report.Id,
		PodName:   report.PodName,
		Namespace: report.Namespace,
		Phase:     report.Phase,
		Reason:    report.Reason,
		Logs:      report.Logs,
	}

	req.ContainerIssues = make([]*archivepb.ContainerIssue, 0, len(report.ContainerIssues))

	for _, ci := range report.ContainerIssues {
		req.ContainerIssues = append(req.ContainerIssues, &archivepb.ContainerIssue{
			Container: ci.Container,
			Reason:    ci.Reason,
		})
	}

	return r.ArchiveClient.Store(ctx, req)
}
