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

package officer

import (
	"context"
	"testing"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// fakeService is a test double for IncidentService.
//
// It records whether HandleIncident was invoked and allows
// simulating service-level errors.
type fakeService struct {
	called bool
	err    error
}

// HandleIncident implements the IncidentService interface.
//
// It marks the method as called and returns a predefined error if set.
func (f *fakeService) HandleIncident(ctx context.Context, r *incident.IncidentReport) error {
	f.called = true
	return f.err
}

// TestReconcile_NoIssue ensures that no incident is generated
// when the Pod is in a healthy state.
func TestReconcile_NoIssue(t *testing.T) {
	pod := &v1.Pod{}

	c := fakeclient.NewClientBuilder().
		WithObjects(pod).
		Build()

	r := &PodReconciler{
		Client:  c,
		Log:     ctrl.Log.WithName("test"),
		Service: &fakeService{},
		Cluster: "test-cluster",
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Service.(*fakeService).called {
		t.Fatal("expected HandleIncident NOT to be called")
	}
}

// TestReconcile_PodNotFound verifies that missing Pods
// are handled gracefully without errors.
func TestReconcile_PodNotFound(t *testing.T) {
	c := fakeclient.NewClientBuilder().Build()

	r := &PodReconciler{
		Client:  c,
		Log:     ctrl.Log.WithName("test"),
		Service: &fakeService{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "missing",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestReconcile_ProblemPod ensures that a failing Pod
// triggers incident creation via the service layer.
func TestReconcile_ProblemPod(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      "bad-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"court.dev/repository": "https://github.com/ghdrope/court",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	}

	c := fakeclient.NewClientBuilder().
		WithObjects(pod).
		Build()

	svc := &fakeService{}

	r := &PodReconciler{
		Client:  c,
		Log:     ctrl.Log.WithName("test"),
		Service: svc,
		Cluster: "test-cluster",
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bad-pod",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !svc.called {
		t.Fatal("expected HandleIncident to be called")
	}
}

// TestReconcile_ServiceFailure verifies that errors from the service layer
// are correctly propagated by the reconciler.
func TestReconcile_ServiceFailure(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      "bad-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"court.dev/repository": "https://github.com/ghdrope/court",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	}

	c := fakeclient.NewClientBuilder().
		WithObjects(pod).
		Build()

	svc := &fakeService{
		err: context.DeadlineExceeded,
	}

	r := &PodReconciler{
		Client:  c,
		Log:     ctrl.Log.WithName("test"),
		Service: svc,
		Cluster: "test-cluster",
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bad-pod",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(context.Background(), req)

	if err == nil {
		t.Fatal("expected error from service")
	}

	if !svc.called {
		t.Fatal("expected HandleIncident to be called")
	}
}
