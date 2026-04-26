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
// It does NOT guarantee absence of issues, only basic health.
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

func shouldIgnorePod(pod *v1.Pod) bool {
	if pod == nil {
		return true
	}

	// NÃO ignorar apenas por idade
	// só ignorar realmente pods que ainda não existem logicamente

	if pod.Status.Phase == v1.PodPending &&
		time.Since(pod.CreationTimestamp.Time) < 5*time.Second {
		return true
	}

	return false
}

func isPodFailing(pod *v1.Pod, issues []incident.ContainerMetadata) bool {
	if pod == nil {
		return false
	}

	age := time.Since(pod.CreationTimestamp.Time)
	isStartup := age < 60*time.Second

	// ======================================================
	// 🔴 1. HARD SIGNAL (sempre válido, inclusive startup)
	// ======================================================
	if len(issues) > 0 {
		return true
	}

	// ======================================================
	// 🔴 2. TERMINATED FAILURES (CRITICAL FIX)
	// ======================================================
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

	// ======================================================
	// 🔴 3. POD PHASE FAILURES (startup-safe)
	// ======================================================
	if pod.Status.Phase == v1.PodFailed {
		return true
	}

	// ======================================================
	// 🟡 4. STARTUP DEGRADED LOGIC (only after grace period)
	// ======================================================
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
//   - It completed successfully (Succeeded)
//   - It finished execution
//   - It is Running and Ready without issues
//
// This represents the "no operational intervention needed" state.
func isPodResolved(pod *v1.Pod, containersMetadata []incident.ContainerMetadata) bool {

	if pod == nil {
		return true
	}

	// If there are active container issues -> not resolved
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
