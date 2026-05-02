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
	"time"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
)

// isPodHealthy checks whether a pod is Running and Ready.
//
// This indicates basic readiness, not absence of issues.
func isPodHealthy(pod *v1.Pod) bool {
	if pod == nil {
		return false
	}

	if pod.Status.Phase != v1.PodRunning {
		return false
	}

	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
			return true
		}
	}

	return false
}

// shouldIgnorePod determines whether a Pod should be skipped during reconciliation.
//
// Pods in very early startup (Pending for a short time) are ignored
// to avoid false positives during scheduling and initialization.
func shouldIgnorePod(pod *v1.Pod) bool {
	if pod == nil {
		return true
	}

	if pod.Status.Phase == v1.PodPending &&
		time.Since(pod.CreationTimestamp.Time) < 5*time.Second {
		return true
	}

	return false
}

// isPodFailing determines whether a Pod is in a failing state.
//
// It combines:
//   - explicit issue detection
//   - container runtime failures
//   - pod phase failures
//   - degraded readiness (after startup grace period)
func isPodFailing(pod *v1.Pod, issues []incident.ContainerMetadata) bool {
	if pod == nil {
		return false
	}

	age := time.Since(pod.CreationTimestamp.Time)
	isStartup := age < 60*time.Second

	// Explicit detected issues (always valid)
	if len(issues) > 0 {
		return true
	}

	// Container-level failures
	for _, cs := range pod.Status.ContainerStatuses {

		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
			return true
		}

		if cs.State.Waiting != nil {
			switch cs.State.Waiting.Reason {
			case "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull", "RunContainerError":
				return true
			}
		}
	}

	// Pod phase failure
	if pod.Status.Phase == v1.PodFailed {
		return true
	}

	// Degraded readiness (after startup grace period)
	if !isStartup {
		if pod.Status.Phase == v1.PodRunning && !isPodHealthy(pod) {
			return true
		}
	}

	return false
}

// isPodResolved determines whether a Pod no longer requires attention.
//
// A Pod is considered resolved when:
//   - it completed successfully (Succeeded)
//   - or it is Running and Ready with no detected issues
func isPodResolved(pod *v1.Pod, containersMetadata []incident.ContainerMetadata) bool {

	if pod == nil {
		return true
	}

	// Active issues -> not resolved
	if len(containersMetadata) > 0 {
		return false
	}

	switch pod.Status.Phase {

	case v1.PodSucceeded:
		return true

	case v1.PodRunning:
		return isPodHealthy(pod)

	default:
		return false
	}
}
