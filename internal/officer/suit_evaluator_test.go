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
	"testing"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestEvaluateSuitClosure verifies the closure decision logic for suits,
// including pod deletion, UID mismatch, and successful resolution states.
func TestEvaluateSuitClosure(t *testing.T) {

	tests := []struct {
		name         string
		pod          *v1.Pod
		expectedUID  string
		metadata     []incident.ContainerMetadata
		expectClose  bool
		expectReason string
	}{
		{
			name:         "pod deleted",
			pod:          nil,
			expectedUID:  "",
			expectClose:  true,
			expectReason: "pod_deleted",
		},
		{
			name: "pod recreated uid mismatch",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID: "new-uid",
				},
			},
			expectedUID:  "old-uid",
			expectClose:  true,
			expectReason: "pod_recreated",
		},
		{
			name: "pod resolved",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID: "same-uid",
				},
				Status: v1.PodStatus{
					Phase: v1.PodSucceeded,
				},
			},
			expectedUID:  "same-uid",
			expectClose:  true,
			expectReason: "pod_resolved",
		},
		{
			name: "pod still active",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID: "same-uid",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
			expectedUID: "same-uid",
			expectClose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			close, reason := EvaluateSuitClosure(
				tt.pod,
				tt.expectedUID,
				tt.metadata,
			)

			if close != tt.expectClose {
				t.Errorf("expected close=%v, got %v", tt.expectClose, close)
			}

			if reason != tt.expectReason {
				t.Errorf("expected reason=%q, got %q", tt.expectReason, reason)
			}
		})
	}
}
