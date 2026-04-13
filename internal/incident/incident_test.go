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
)

// TestIncidentReport_Structure ensures that the IncidentReport structure
// correctly holds and exposes data without transformation or side effects.
func TestIncidentReport_Structure(t *testing.T) {
	report := IncidentReport{
		ID: "test-id",

		Cluster:   "cluster-1",
		Namespace: "default",
		Pod:       "pod-1",

		Events: []K8sEvent{
			{
				Type:    "Normal",
				Reason:  "BackOff",
				Message: "Back-off restarting failed container",
			},
		},

		ContainerIssues: []ContainerIssue{
			{
				Container: "app",
				Reason:    "CrashLoopBackOff",
				Logs:      []string{"Back-off pulling image app/latest"},
			},
		},
	}

	// Validate identity
	if report.ID != "test-id" {
		t.Errorf("unexpected ID: got %s, want %s", report.ID, "test-id")
	}

	if report.Cluster != "cluster-1" {
		t.Errorf("unexpected ID: got %s, want %s", report.Cluster, "cluster-1")
	}

	if report.Pod != "pod-1" {
		t.Errorf("unexpected pod: got %s, want %s", report.Pod, "pod-1")
	}

	if report.Namespace != "default" {
		t.Errorf("unexpected namespace: got %s, want %s", report.Namespace, "default")
	}

	// Validate events
	if len(report.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(report.Events))
	}

	if report.Events[0].Reason != "BackOff" {
		t.Errorf("unexpected event reason: got %s, want %s", report.Events[0].Reason, "BackOff")
	}

	// Validate container issues
	if len(report.ContainerIssues) != 1 {
		t.Fatalf("expected 1 container issue, got %d", len(report.ContainerIssues))
	}

	if report.ContainerIssues[0].Reason != "CrashLoopBackOff" {
		t.Errorf("unexpected container issue reason: got %s, want %s",
			report.ContainerIssues[0].Reason,
			"CrashLoopBackOff",
		)
	}

	// Validate logs inside container issue
	if len(report.ContainerIssues[0].Logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(report.ContainerIssues[0].Logs))
	}

	if report.ContainerIssues[0].Logs[0] != "Back-off pulling image app/latest" {
		t.Errorf("unexpected log content: got %s, want %s",
			report.ContainerIssues[0].Logs[0],
			"Back-off pulling image app/latest",
		)
	}
}

// TestContainerIssue_Structure ensures that ContainerIssue behaves
// as a simple data container without hidden logic.
func TestContainerIssue_Structure(t *testing.T) {
	ci := ContainerIssue{
		Container: "app",
		Reason:    "OOMKilled",
		Logs:      []string{"out of memory"},
	}

	if ci.Container != "app" {
		t.Errorf("unexpected container: got %s, want %s", ci.Container, "nginx")
	}

	if ci.Reason != "OOMKilled" {
		t.Errorf("unexpected reason: got %s, want %s", ci.Reason, "OOMKilled")
	}

	if len(ci.Logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(ci.Logs))
	}
}
