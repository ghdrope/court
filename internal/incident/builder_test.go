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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestBuildFromPod ensures a Kubernetes Pod is correctly converted
// into an IncidentReport domain model.
func TestBuildFromPod(t *testing.T) {
	t.Parallel()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api",
			Namespace: "ns-1",
			UID:       "uid-123",
		},
	}

	events := []K8sEvent{
		{Type: "Normal", Reason: "Started", Message: "ok"},
	}

	containers := []ContainerMetadata{
		{
			Container: "api",
			ImageName: "nginx",
			Reason:    "Crash",
			Logs:      []string{"failed"},
		},
	}

	inc, err := BuildFromPod(
		pod,
		"cluster-1",
		"https://github.com/example",
		events,
		containers,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inc.Pod != "api" {
		t.Errorf("expected api, got %s", inc.Pod)
	}

	if inc.Namespace != "ns-1" {
		t.Errorf("expected ns-1, got %s", inc.Namespace)
	}

	if len(inc.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(inc.Events))
	}
}

// TestBuildFromPod_Nil ensures nil pod returns error.
func TestBuildFromPod_Nil(t *testing.T) {
	t.Parallel()

	_, err := BuildFromPod(nil, "c1", "", nil, nil)

	if err == nil {
		t.Errorf("expected error for nil pod")
	}
}
