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

package incident

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

// BuildFromPod constructs an IncidentReport from a K8s Pod.
//
// It acts as a translation layer between K8s runtime state
// and the domain IncidentReport model.
func BuildFromPod(
	pod *v1.Pod,
	cluster string,
	repoURL string,
	events []K8sEvent,
	containerIssues []ContainerIssue,
) (IncidentReport, error) {

	if pod == nil {
		return IncidentReport{}, fmt.Errorf("pod is nil")
	}

	return IncidentReport{
		// ID is derived from namespace/name/UID to guarantee uniqueness over time.
		ID: fmt.Sprintf("%s/%s/%s", pod.Namespace, pod.Name, pod.UID),

		Cluster:   cluster,
		Namespace: pod.Namespace,
		Pod:       pod.Name,

		VCSRepoURL: repoURL,

		// Core evidence
		Events: events,

		// Container-level evidence
		ContainerIssues: containerIssues,
	}, nil
}
