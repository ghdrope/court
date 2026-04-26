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
	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
)

// EvaluateSuitClosure determines whether a suit should be closed.
//
// This is a no-side effects function shared by both recovery and controller.
//
// expectedUID:
//   - recovery: UID from stored incident
//   - controller: empty string (not applicable)
func EvaluateSuitClosure(
	pod *v1.Pod,
	expectedUID string,
	containerIssues []incident.ContainerIssue,
) (bool, string) {

	// Pod no longer exists in cluster
	if pod == nil {
		return true, "pod_deleted"
	}

	// Pod was recreated (new UID)
	if expectedUID != "" && string(pod.UID) != expectedUID {
		return true, "pod_recreated"
	}

	// Pod recovery condition
	if isPodResolved(pod, containerIssues) {
		return true, "pod_resolved"
	}

	return false, ""
}
