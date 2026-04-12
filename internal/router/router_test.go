package router

import (
	"context"
	"errors"
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	archivepb "github.com/ghdrope/court/proto/archive"
	pb "github.com/ghdrope/court/proto/incident"
)

// TestRouter_Route_Success ensures correct transformation and successful forwarding.
func TestRouter_Route_Success(t *testing.T) {
	called := false

	mock := &testhelper.MockArchiveClient{
		StoreFunc: func(ctx context.Context, req *archivepb.StoreIncidentRequest) error {
			called = true

			// Validate mapping from IncidentReport → StoreIncidentRequest
			if req.Id != "id-1" {
				t.Errorf("unexpected id: %s", req.Id)
			}

			if req.PodName != "pod-1" {
				t.Errorf("unexpected pod name: %s", req.PodName)
			}

			if req.Namespace != "default" {
				t.Errorf("unexpected namespace: %s", req.Namespace)
			}

			if req.Phase != "Failed" {
				t.Errorf("unexpected phase: %s", req.Phase)
			}

			if req.Reason != "CrashLoopBackOff" {
				t.Errorf("unexpected reason: %s", req.Reason)
			}

			if len(req.ContainerIssues) != 1 {
				t.Fatalf("expected 1 container issue, got %d", len(req.ContainerIssues))
			}

			if req.ContainerIssues[0].Container != "app" {
				t.Errorf("unexpected container: %s", req.ContainerIssues[0].Container)
			}

			if len(req.Logs) != 2 {
				t.Fatalf("expected 2 logs, got %d", len(req.Logs))
			}

			return nil
		},
	}

	r := &Router{
		ArchiveClient: mock,
	}

	report := &pb.IncidentReport{
		Id:        "id-1",
		PodName:   "pod-1",
		Namespace: "default",
		Phase:     "Failed",
		Reason:    "CrashLoopBackOff",
		ContainerIssues: []*pb.ContainerIssue{
			{
				Container: "app",
				Reason:    "CrashLoopBackOff",
			},
		},
		Logs: []string{"log1", "log2"},
	}

	err := r.Route(context.Background(), report)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("expected ArchiveClient.Store to be called")
	}
}

//
// ---- Test: nil input ----
//

// TestRouter_Route_NilReport ensures nil report is handled correctly.
func TestRouter_Route_NilReport(t *testing.T) {
	r := &Router{
		ArchiveClient: &testhelper.MockArchiveClient{},
	}

	err := r.Route(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for nil report")
	}
}

// TestRouter_Route_ArchiveError ensures errors from archive service are propagated.
func TestRouter_Route_ArchiveError(t *testing.T) {
	mock := &testhelper.MockArchiveClient{
		StoreFunc: func(ctx context.Context, req *archivepb.StoreIncidentRequest) error {
			return errors.New("archive failure")
		},
	}

	r := &Router{
		ArchiveClient: mock,
	}

	err := r.Route(context.Background(), &pb.IncidentReport{
		Id:        "id-1",
		PodName:   "pod-1",
		Namespace: "default",
	})

	if err == nil {
		t.Fatal("expected error")
	}
}
