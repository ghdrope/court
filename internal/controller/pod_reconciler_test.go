package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	pb "github.com/ghdrope/court/proto/incident"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestReconcile_NoProblem ensures no incident is sent when pod is healthy.
func TestReconcile_NoProblem(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	pod := testhelper.NewTestPod("default", "pod-ok", v1.PodRunning)

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(pod).
		Build()

	apiCalled := false

	reconciler := &PodReconciler{
		Client: client,
		Log:    logr.Discard(),
		API: &testhelper.MockAPIClient{
			SendFunc: func(ctx context.Context, report *pb.IncidentReport) error {
				apiCalled = true
				return nil
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "pod-ok",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if apiCalled {
		t.Fatal("expected API not to be called for healthy pod")
	}
}

// TestReconcile_FailedPod ensures incident is sent when pod is failed.
func TestReconcile_FailedPod(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	pod := testhelper.NewTestPod("default", "pod-failed", v1.PodFailed)

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(pod).
		Build()

	apiCalled := false

	reconciler := &PodReconciler{
		Client: client,
		Log:    logr.Discard(),
		API: &testhelper.MockAPIClient{
			SendFunc: func(ctx context.Context, report *pb.IncidentReport) error {
				apiCalled = true

				// Validate important fields
				if report.PodName != "pod-failed" {
					t.Errorf("unexpected pod name: %s", report.PodName)
				}

				return nil
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "pod-failed",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !apiCalled {
		t.Fatal("expected API to be called")
	}
}

// TestReconcile_APIFailure ensures errors from API are propagated.
func TestReconcile_APIFailure(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	pod := testhelper.NewTestPod("default", "pod-failed", v1.PodFailed)

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(pod).
		Build()

	reconciler := &PodReconciler{
		Client: client,
		Log:    logr.Discard(),
		API: &testhelper.MockAPIClient{
			SendFunc: func(ctx context.Context, report *pb.IncidentReport) error {
				return errors.New("api failure")
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "pod-failed",
		},
	})

	if err == nil {
		t.Fatal("expected error from API")
	}
}

// TestReconcile_PodNotFound ensures not found pods are ignored.
func TestReconcile_PodNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	reconciler := &PodReconciler{
		Client: client,
		Log:    logr.Discard(),
		API: &testhelper.MockAPIClient{
			SendFunc: func(ctx context.Context, report *pb.IncidentReport) error {
				t.Fatal("API should not be called")
				return nil
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "does-not-exist",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
