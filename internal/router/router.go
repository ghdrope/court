package router

import (
	"context"
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

// Route convers and forwards the incident to Archive.
func (r *Router) Route(ctx context.Context, report *pb.IncidentReport) error {

	log.Printf("routing incident to archive: %s", report.EventId)

	req := &archivepb.StoreIncidentRequest{
		EventId:   report.EventId,
		PodName:   report.PodName,
		Namespace: report.Namespace,
		Phase:     report.Phase,
		Reason:    report.Reason,
		Logs:      report.Logs,
	}

	// map container issues
	for _, ci := range report.ContainerIssues {
		req.ContainerIssues = append(req.ContainerIssues, &archivepb.ContainerIssue{
			Container: ci.Container,
			Reason:    ci.Reason,
		})
	}

	return r.ArchiveClient.Store(ctx, req)
}
