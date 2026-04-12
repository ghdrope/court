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

package inspector

import (
	"testing"

	"github.com/ghdrope/court/pkg/testhelper"
	v1 "k8s.io/api/core/v1"
)

// TestDetectContainerIssues_NilPod ensures nil pod returns nil.
func TestDetectContainerIssues_NilPod(t *testing.T) {
	issues := DetectContainerIssues(nil)

	if issues != nil {
		t.Fatal("expected nil issues for nil pod")
	}
}

// TestDetectContainerIssues_NoIssues ensures no issues are returned for healthy containers.
func TestDetectContainerIssues_NoIssues(t *testing.T) {
	pod := testhelper.NewPodWithStatuses([]v1.ContainerStatus{
		{
			Name:  "app",
			State: v1.ContainerState{}, // no waiting or terminated state
		},
	})

	issues := DetectContainerIssues(pod)

	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

// TestDetectContainerIssues_WaitingStates validates detection of waiting issues.
func TestDetectContainerIssues_WaitingStates(t *testing.T) {
	pod := testhelper.NewPodWithStatuses([]v1.ContainerStatus{
		testhelper.NewContainerStatusWaiting("app", "CrashLoopBackOff"),
		testhelper.NewContainerStatusWaiting("sidecar", "ImagePullBackOff"),
	})

	issues := DetectContainerIssues(pod)

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}

	expected := map[string]string{
		"app":     "CrashLoopBackOff",
		"sidecar": "ImagePullBackOff",
	}

	for _, issue := range issues {
		if expected[issue.Container] != issue.Reason {
			t.Errorf("unexpected issue: %+v", issue)
		}
	}
}

// TestDetectContainerIssues_TerminatedOOM validates detection of OOMKilled containers.
func TestDetectContainerIssues_TerminatedOOM(t *testing.T) {
	pod := testhelper.NewPodWithStatuses([]v1.ContainerStatus{
		testhelper.NewContainerStatusTerminated("app", "OOMKilled"),
	})

	issues := DetectContainerIssues(pod)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	if issues[0].Reason != "OOMKilled" {
		t.Errorf("unexpected reason: %s", issues[0].Reason)
	}
}

// TestDetectContainerIssues_MixedStates validates multiple conditions together.
func TestDetectContainerIssues_MixedStates(t *testing.T) {
	pod := testhelper.NewPodWithStatuses([]v1.ContainerStatus{
		testhelper.NewContainerStatusWaiting("app", "CrashLoopBackOff"),
		testhelper.NewContainerStatusTerminated("worker", "OOMKilled"),
		testhelper.NewContainerStatusWaiting("init", "ImagePullBackOff"),
	})

	issues := DetectContainerIssues(pod)

	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(issues))
	}
}

// TestDetectContainerIssues_IgnoresIrrelevantReasons ensures unrelated reasons are ignored.
func TestDetectContainerIssues_IgnoresIrrelevantReasons(t *testing.T) {
	pod := testhelper.NewPodWithStatuses([]v1.ContainerStatus{
		testhelper.NewContainerStatusWaiting("app", "ContainerCreating"),
		testhelper.NewContainerStatusTerminated("worker", "Completed"),
	})

	issues := DetectContainerIssues(pod)

	if len(issues) != 0 {
		t.Fatalf("expected 0 issues, got %d", len(issues))
	}
}
