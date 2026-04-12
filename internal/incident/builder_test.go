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
	v1 "k8s.io/api/core/v1"
)

// TestBuildFromPod_Success validates that a valid pod produces a correct IncidentReport.
func TestBuildFromPod_Success(t *testing.T) {
	pod := testhelper.NewTestPod("default", "pod-1", v1.PodFailed)

	containerIssues := []ContainerIssue{
		{Container: "app", Reason: "CrashLoopBackOff"},
	}

	logs := []string{"log1", "log2"}

	report, err := BuildFromPod(pod, containerIssues, logs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate generated ID
	if report.ID == "" {
		t.Error("expected non-empty ID")
	}

	// Validate basic fields mapping
	if report.PodName != "pod-1" {
		t.Errorf("unexpected pod name: %s", report.PodName)
	}

	if report.Namespace != "default" {
		t.Errorf("unexpected namespace: %s", report.Namespace)
	}

	if report.Phase != v1.PodFailed {
		t.Errorf("unexpected phase: %s", report.Phase)
	}

	// Reason is derived from phase
	if report.Reason != string(v1.PodFailed) {
		t.Errorf("unexpected reason: %s", report.Reason)
	}

	// Validate container issues propagation
	if len(report.ContainerIssues) != 1 {
		t.Fatalf("expected 1 container issue, got %d", len(report.ContainerIssues))
	}

	if report.ContainerIssues[0].Container != "app" {
		t.Errorf("unexpected container: %s", report.ContainerIssues[0].Container)
	}

	// Validate logs propagation
	if len(report.Logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(report.Logs))
	}
}

// TestBuildFromPod_NilPod ensures nil pod returns an error.
func TestBuildFromPod_NilPod(t *testing.T) {
	report, err := BuildFromPod(nil, nil, nil)

	if err == nil {
		t.Fatal("expected error for nil pod")
	}

	if report.ID != "" {
		t.Error("expected empty report on error")
	}
}

// TestBuildFromPod_EmptySlices ensures empty inputs are handled correctly.
func TestBuildFromPod_EmptySlices(t *testing.T) {
	pod := testhelper.NewTestPod("default", "pod-empty", v1.PodRunning)

	report, err := BuildFromPod(pod, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Slices should remain nil
	if report.ContainerIssues != nil {
		t.Error("expected nil container issues")
	}

	if report.Logs != nil {
		t.Error("expected nil logs")
	}
}

// TestBuildFromPod_TableDriven validates multiple pod phases.
func TestBuildFromPod_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		phase v1.PodPhase
	}{
		{"Running", v1.PodRunning},
		{"Failed", v1.PodFailed},
		{"Pending", v1.PodPending},
		{"Unknown", v1.PodUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := testhelper.NewTestPod("ns", "pod", tt.phase)

			report, err := BuildFromPod(pod, nil, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if report.Phase != tt.phase {
				t.Errorf("expected phase %s, got %s", tt.phase, report.Phase)
			}

			if report.Reason != string(tt.phase) {
				t.Errorf("expected reason %s, got %s", tt.phase, report.Reason)
			}
		})
	}
}
