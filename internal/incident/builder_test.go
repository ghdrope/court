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

package incident

import (
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
)

// TestBuildFromPod_Success validates that a valid pod produces a correct
// IncidentReport with proper identity and container issue propagation.
func TestBuildFromPod_Success(t *testing.T) {
	pod := testhelper.NewTestPod("default", "pod-1")

	cluster := "test-cluster"

	events := []K8sEvent{
		{
			Type:    "Normal",
			Reason:  "BackOff",
			Message: "Back-off restarting failed container",
		},
	}

	containerIssues := []ContainerIssue{
		{
			Container: "app",
			Reason:    "CrashLoopBackOff",
			Logs:      []string{"log1", "log2"},
		},
	}

	report, err := BuildFromPod(pod, cluster, events, containerIssues)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate ID generation
	if report.ID == "" {
		t.Error("expected non-empty ID")
	}

	if report.Cluster != cluster {
		t.Errorf("expected cluster %s, got %s", cluster, report.Cluster)
	}

	// Validate identity mapping
	if report.Pod != "pod-1" {
		t.Errorf("unexpected pod: got %s, want %s", report.Pod, "pod-1")
	}

	if report.Namespace != "default" {
		t.Errorf("unexpected namespace: got %s, want %s", report.Namespace, "default")
	}

	// Validate events propagation
	if len(report.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(report.Events))
	}

	if report.Events[0].Reason != "BackOff" {
		t.Errorf("unexpected event reason: got %s, want %s",
			report.Events[0].Reason,
			"BackOff",
		)
	}

	// Validate container issues propagation
	if len(report.ContainerIssues) != 1 {
		t.Fatalf("expected 1 container issue, got %d", len(report.ContainerIssues))
	}

	if report.ContainerIssues[0].Container != "app" {
		t.Errorf("unexpected container: got %s, want %s",
			report.ContainerIssues[0].Container,
			"app",
		)
	}

	if report.ContainerIssues[0].Reason != "CrashLoopBackOff" {
		t.Errorf("unexpected reason: got %s, want %s",
			report.ContainerIssues[0].Reason,
			"CrashLoopBackOff",
		)
	}

	// Validate logs inside container issue
	if len(report.ContainerIssues[0].Logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(report.ContainerIssues[0].Logs))
	}
}

// TestBuildFromPod_NilPod ensures nil pod input returns an error
// and does not produce a valid IncidentReport.
func TestBuildFromPod_NilPod(t *testing.T) {
	report, err := BuildFromPod(nil, "", nil, nil)

	if err == nil {
		t.Fatal("expected error for nil pod")
	}

	if report.ID != "" {
		t.Error("expected empty report on error")
	}
}

// TestBuildFromPod_EmptySlices ensures empty that empty inputs are handled
// correctly without causing nil pointer issues or unexpected data.
func TestBuildFromPod_EmptySlices(t *testing.T) {
	pod := testhelper.NewTestPod("default", "pod-empty")

	report, err := BuildFromPod(pod, "cluster", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Events should be nil when not provided
	if report.Events != nil {
		t.Error("expected nil events slice")
	}

	// Container issues should be nil when not provided
	if report.ContainerIssues != nil {
		t.Error("expected nil container issues slice")
	}
}

// TestBuildFromPod_TableDriven validates multiple pod identities
// are correctly mapped into IncidentReport without depending on phase logic.
func TestBuildFromPod_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		podName   string
	}{
		{
			name:      "basic mapping",
			namespace: "ns-1",
			podName:   "pod-a",
		},
		{
			name:      "another pod",
			namespace: "ns-2",
			podName:   "pod-b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := testhelper.NewTestPod(tt.namespace, tt.podName)

			report, err := BuildFromPod(pod, "cluster", nil, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if report.Namespace != tt.namespace {
				t.Errorf("expected namespace %s, got %s",
					tt.namespace,
					report.Namespace,
				)
			}

			if report.Pod != tt.podName {
				t.Errorf("expected pod %s, got %s",
					tt.podName,
					report.Pod,
				)
			}
		})
	}
}
