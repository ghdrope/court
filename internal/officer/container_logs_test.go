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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestEnrichContainersMetadataWithLogs_AddsMissingContainers verifies that
// containers without detected issues are still included in the result set.
func TestEnrichContainersMetadataWithLogs_AddsMissingContainers(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-pod",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "app", Image: "busybox"},
			},
		},
	}

	// client intentionally nil
	issues := []incident.ContainerMetadata{}

	result := EnrichContainersMetadataWithLogs(
		context.Background(),
		nil, // client intentionally nil
		pod,
		issues,
	)

	if len(result) != 1 {
		t.Fatalf("expected 1 container metadata, got %d", len(result))
	}

	if result[0].Container != "app" {
		t.Fatalf("unexpected container name")
	}
}
