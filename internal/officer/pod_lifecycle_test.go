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
	"time"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestIsPodHealthy verifies that a Pod is considered healthy only when it is
// in Running phase and has a Ready condition set to True.
func TestIsPodHealthy(t *testing.T) {

	tests := []struct {
		name string
		pod  *v1.Pod
		want bool
	}{
		{
			name: "nil pod",
			pod:  nil,
			want: false,
		},
		{
			name: "not running",
			pod: &v1.Pod{
				Status: v1.PodStatus{Phase: v1.PodPending},
			},
			want: false,
		},
		{
			name: "running but not ready",
			pod: &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
			want: false,
		},
		{
			name: "running and ready",
			pod: &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					Conditions: []v1.PodCondition{
						{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPodHealthy(tt.pod)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestShouldIgnorePod verifies that Pods in early Pending state are ignored
// to avoid false positives during startup and scheduling.
func TestShouldIgnorePod(t *testing.T) {

	tests := []struct {
		name string
		pod  *v1.Pod
		want bool
	}{
		{
			name: "nil pod ignored",
			pod:  nil,
			want: true,
		},
		{
			name: "recent pending pod ignored",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
			want: true,
		},
		{
			name: "old pending pod not ignored",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-10 * time.Second)),
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIgnorePod(tt.pod)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestIsPodFailing verifies detection of failing Pods based on container state,
// pod phase, explicit issues, and degraded readiness after startup grace period.
func TestIsPodFailing(t *testing.T) {

	tests := []struct {
		name       string
		pod        *v1.Pod
		issues     []incident.ContainerMetadata
		expectFail bool
	}{
		{
			name:       "nil pod",
			pod:        nil,
			expectFail: false,
		},
		{
			name: "explicit issues",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
			},
			issues:     []incident.ContainerMetadata{{}},
			expectFail: true,
		},
		{
			name: "failed pod phase",
			pod: &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodFailed,
				},
			},
			expectFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPodFailing(tt.pod, tt.issues)
			if got != tt.expectFail {
				t.Errorf("expected %v, got %v", tt.expectFail, got)
			}
		})
	}
}

// TestIsPodResolved verifies that a Pod is considered resolved when it has no
// active issues and is either succeeded or running and healthy.
func TestIsPodResolved(t *testing.T) {

	tests := []struct {
		name   string
		pod    *v1.Pod
		issues []incident.ContainerMetadata
		want   bool
	}{
		{
			name: "nil pod resolved",
			pod:  nil,
			want: true,
		},
		{
			name: "has issues not resolved",
			pod: &v1.Pod{
				Status: v1.PodStatus{Phase: v1.PodRunning},
			},
			issues: []incident.ContainerMetadata{{}},
			want:   false,
		},
		{
			name: "succeeded pod",
			pod: &v1.Pod{
				Status: v1.PodStatus{Phase: v1.PodSucceeded},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPodResolved(tt.pod, tt.issues)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
