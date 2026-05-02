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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestFetchPodEvents verifies that fetchPodEvents correctly filters Kubernetes
// events belonging only to the target Pod, matching by kind, UID, and name,
// and ignores unrelated or empty events.
func TestFetchPodEvents(t *testing.T) {

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			UID:       "12345",
		},
	}

	tests := []struct {
		name      string
		events    []v1.Event
		expectLen int
	}{
		{
			name: "filters matching pod events only",
			events: []v1.Event{
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "test-pod",
						UID:  "12345",
					},
					Type:    "Warning",
					Reason:  "Failed",
					Message: "something failed",
				},
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Pod",
						Name: "other-pod",
						UID:  "999",
					},
					Type: "Warning",
				},
				{
					InvolvedObject: v1.ObjectReference{
						Kind: "Node",
					},
					Type: "Warning",
				},
			},
			expectLen: 1,
		},
		{
			name:      "no events returns empty slice",
			events:    []v1.Event{},
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := fake.NewSimpleClientset()

			for _, e := range tt.events {
				_, _ = client.CoreV1().
					Events("default").
					Create(context.TODO(), &e, metav1.CreateOptions{})
			}

			r := &PodReconciler{
				KubeClient: client,
			}

			res, err := r.fetchPodEvents(context.TODO(), "default", pod)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(res) != tt.expectLen {
				t.Errorf("expected %d events, got %d", tt.expectLen, len(res))
			}

			for _, ev := range res {
				if ev.Type == "" && ev.Reason == "" && ev.Message == "" {
					t.Error("empty event should not be returned")
				}
			}
		})
	}
}
