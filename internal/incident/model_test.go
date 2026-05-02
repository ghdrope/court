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

import "testing"

// TestIncidentReport_Structure ensures the IncidentReport model
// correctly stores all fields without mutation or loss of data.
func TestIncidentReport_Structure(t *testing.T) {
	t.Parallel()

	inc := IncidentReport{
		ID:         "ns-1/api/uid-123",
		Cluster:    "cluster-1",
		Namespace:  "ns-1",
		Pod:        "api",
		VCSRepoURL: "https://github.com/example/repo",
		Events: []K8sEvent{
			{
				Type:    "Normal",
				Reason:  "Started",
				Message: "container started",
			},
		},
		ContainersMetadata: []ContainerMetadata{
			{
				Container: "api",
				ImageName: "nginx:latest",
				Reason:    "CrashLoop",
				Logs:      []string{"error starting server"},
			},
		},
	}

	if inc.ID != "ns-1/api/uid-123" {
		t.Errorf("expected ID ns-1/api/uid-123, got %s", inc.ID)
	}

	if inc.Cluster != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", inc.Cluster)
	}

	if inc.Namespace != "ns-1" {
		t.Errorf("expected ns-1, got %s", inc.Namespace)
	}

	if inc.Pod != "api" {
		t.Errorf("expected api, got %s", inc.Pod)
	}

	if inc.VCSRepoURL != "https://github.com/example/repo" {
		t.Errorf("unexpected VCSRepoURL: %s", inc.VCSRepoURL)
	}

	if len(inc.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(inc.Events))
	}

	if inc.Events[0].Type != "Normal" {
		t.Errorf("expected event type Normal, got %s", inc.Events[0].Type)
	}

	if len(inc.ContainersMetadata) != 1 {
		t.Fatalf("expected 1 container metadata, got %d", len(inc.ContainersMetadata))
	}

	if inc.ContainersMetadata[0].Container != "api" {
		t.Errorf("expected container api, got %s", inc.ContainersMetadata[0].Container)
	}
}

// TestK8sEvent_Structure ensures event fields are correctly stored.
func TestK8sEvent_Structure(t *testing.T) {
	t.Parallel()

	event := K8sEvent{
		Type:    "Warning",
		Reason:  "OOMKilled",
		Message: "container exceeded memory limit",
	}

	if event.Type != "Warning" {
		t.Errorf("expected Warning, got %s", event.Type)
	}

	if event.Reason != "OOMKilled" {
		t.Errorf("expected OOMKilled, got %s", event.Reason)
	}

	if event.Message != "container exceeded memory limit" {
		t.Errorf("unexpected message: %s", event.Message)
	}
}

// TestContainerMetadata_Structure ensures container metadata is stored correctly.
func TestContainerMetadata_Structure(t *testing.T) {
	t.Parallel()

	meta := ContainerMetadata{
		Container: "worker",
		ImageName: "redis:7",
		Reason:    "CrashLoopBackOff",
		Logs:      []string{"connection refused", "retrying"},
	}

	if meta.Container != "worker" {
		t.Errorf("expected worker, got %s", meta.Container)
	}

	if meta.ImageName != "redis:7" {
		t.Errorf("expected redis:7, got %s", meta.ImageName)
	}

	if meta.Reason != "CrashLoopBackOff" {
		t.Errorf("expected CrashLoopBackOff, got %s", meta.Reason)
	}

	if len(meta.Logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(meta.Logs))
	}
}
