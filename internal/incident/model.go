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

// IncidentReport is a representation of a Kubernetes workload failure.
//
// It is produced ONLY for workloads whose failure type has been recognized
// by the pod reconciler.
type IncidentReport struct {
	// ID uniquely identifies the incident.
	ID string

	// Target workload identity.
	Cluster   string
	Namespace string
	Pod       string

	// VCSRepoURL represents the source repository of the workload.
	// It is resolved from annotations or image metadata.
	VCSRepoURL string

	// Events contains normalized Kubernetes events related to the failure.
	Events []K8sEvent

	// ContainerIssues contains failure signals for affected containers.
	ContainerIssues []ContainerIssue
}

// K8sEvent represents a single Kubernetes event associated with a Pod.
//
// This is a normalized version of kubectl describe output.
//
// Example transformation:
//
//	Normal  Pulling  Pulling image "app/latest"
//	↓
//	K8sEvent{
//	    Type: "Normal",
//	    Reason: "Pulling",
//	    Message: "Pulling image \"app/latest\"",
//	}
type K8sEvent struct {
	Type    string
	Reason  string
	Message string
}

// ContainerIssue represents a failure observed in a specific container.
type ContainerIssue struct {
	Container string
	ImageName string
	Reason    string
	Logs      []string
}
