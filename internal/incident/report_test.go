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

	v1 "k8s.io/api/core/v1"
)

// TestIncidentReport_Structure ensures the struct fields behave as expected.
func TestIncidentReport_Structure(t *testing.T) {
	report := IncidentReport{
		ID:        "test-id",
		PodName:   "pod-1",
		Namespace: "default",
		Phase:     v1.PodFailed,
		Reason:    "Failed",
		ContainerIssues: []ContainerIssue{
			{Container: "app", Reason: "CrashLoopBackOff"},
		},
		Logs: []string{"log1"},
	}

	// Validate basic fields
	if report.ID != "test-id" {
		t.Errorf("unexpected ID: %s", report.ID)
	}

	if report.PodName != "pod-1" {
		t.Errorf("unexpected pod name: %s", report.PodName)
	}

	// Validate container issues
	if len(report.ContainerIssues) != 1 {
		t.Fatalf("expected 1 container issue")
	}

	if report.ContainerIssues[0].Reason != "CrashLoopBackOff" {
		t.Errorf("unexpected container issue reason")
	}

	// Validate logs
	if len(report.Logs) != 1 {
		t.Fatalf("expected 1 log entry")
	}
}

// TestContainerIssue_Structure ensures ContainerIssue struct correctness.
func TestContainerIssue_Structure(t *testing.T) {
	ci := ContainerIssue{
		Container: "nginx",
		Reason:    "OOMKilled",
	}

	if ci.Container != "nginx" {
		t.Errorf("unexpected container: %s", ci.Container)
	}

	if ci.Reason != "OOMKilled" {
		t.Errorf("unexpected reason: %s", ci.Reason)
	}
}
