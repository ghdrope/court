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
	"fmt"
	"testing"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

// TestFetchPodEvents verifies that only valid Pod events are returned
// and properly mapped into the internal incident.K8sEvent model.
func TestFetchPodEvents(t *testing.T) {
	ctx := context.Background()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			UID:       types.UID("pod-uid-123"),
		},
	}

	tests := []struct {
		name           string
		events         []v1.Event
		expectedCount  int
		expectedEvents []incident.K8sEvent
	}{
		{
			name: "valid pod events are returned",
			events: []v1.Event{
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "test-pod",
						UID:  types.UID("pod-uid-123"),
					},
					Type:    "Warning",
					Reason:  "Failed",
					Message: "Something failed",
				},
			},
			expectedCount: 1,
			expectedEvents: []incident.K8sEvent{
				{
					Type:    "Warning",
					Reason:  "Failed",
					Message: "Something failed",
				},
			},
		},
		{
			name: "non-pod events are ignored",
			events: []v1.Event{
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Node",
					},
					Type:    "Normal",
					Reason:  "Scheduled",
					Message: "Node scheduled",
				},
			},
			expectedCount: 0,
		},
		{
			name: "different UID is ignored",
			events: []v1.Event{
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "test-pod",
						UID:  types.UID("other-uid"),
					},
					Type:    "Warning",
					Reason:  "Failed",
					Message: "Mismatch UID",
				},
			},
			expectedCount: 0,
		},
		{
			name: "different name is ignored",
			events: []v1.Event{
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "other-pod",
						UID:  types.UID("pod-uid-123"),
					},
					Type:    "Warning",
					Reason:  "Failed",
					Message: "Mismatch name",
				},
			},
			expectedCount: 0,
		},
		{
			name: "empty events are ignored",
			events: []v1.Event{
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "test-pod",
						UID:  types.UID("pod-uid-123"),
					},
					Type:    "",
					Reason:  "",
					Message: "",
				},
			},
			expectedCount: 0,
		},
		{
			name: "mixed events only valid are returned",
			events: []v1.Event{
				// valid
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "test-pod",
						UID:  types.UID("pod-uid-123"),
					},
					Type:    "Warning",
					Reason:  "CrashLoopBackOff",
					Message: "Container crashed",
				},
				// wrong kind
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Node",
					},
				},
				// empty
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "test-pod",
						UID:  types.UID("pod-uid-123"),
					},
				},
			},
			expectedCount: 1,
			expectedEvents: []incident.K8sEvent{
				{
					Type:    "Warning",
					Reason:  "CrashLoopBackOff",
					Message: "Container crashed",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			// Seed events into fake client
			for i, e := range tt.events {
				e.Namespace = "default"
				e.Name = fmt.Sprintf("event-%d", i)

				_, err := client.CoreV1().
					Events("default").
					Create(ctx, &e, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create fake event: %v", err)
				}
			}

			r := &PodReconciler{
				KubeClient: client,
			}

			result, err := r.fetchPodEvents(ctx, "default", pod)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tt.expectedCount {
				t.Fatalf("expected %d events, got %d", tt.expectedCount, len(result))
			}

			for i := range tt.expectedEvents {
				if result[i] != tt.expectedEvents[i] {
					t.Errorf("expected event %+v, got %+v", tt.expectedEvents[i], result[i])
				}
			}
		})
	}
}

// TestFetchPodEvents_NilClient verifies defensive behavior when KubeClient is nil.
func TestFetchPodEvents_NilClient(t *testing.T) {
	r := &PodReconciler{
		KubeClient: nil,
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			UID:  types.UID("uid"),
		},
	}

	result, err := r.fetchPodEvents(context.Background(), "default", pod)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
