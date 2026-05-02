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
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestDetectContainerIssues_TerminatedFailure verifies that terminated containers
// with non-zero exit codes are correctly detected as failures.
func TestDetectContainerIssues_TerminatedFailure(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Minute)),
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:  "app",
					Image: "busybox",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 1,
							Reason:   "Error",
						},
					},
				},
			},
		},
	}

	issues := DetectContainersMetadata(context.Background(), nil, pod)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	if issues[0].Container != "app" {
		t.Fatalf("unexpected container name: %s", issues[0].Container)
	}
}

// TestDetectContainerIssues_TerminatedFailure verifies that terminated containers
// with non-zero exit codes are correctly detected as failures.
func TestDetectContainerIssues_IgnoresYoungPods(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
	}

	issues := DetectContainersMetadata(context.Background(), nil, pod)

	if issues != nil {
		t.Fatalf("expected nil issues for young pod")
	}
}
