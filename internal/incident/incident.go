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
	// Identity uniquely identifies the incident.
	ID string

	// Target workload identity
	Cluster   string
	Namespace string
	Pod       string

	// Core diagnostic evidence extracted from Kubernetes events.
	Events []K8sEvent

	// ContainerIssues contains all containers in the Pod that are
	// associated with the recognized failure condition.
	//
	// Each entry includes only relevant error state + logs.
	ContainerIssues []ContainerIssue

	// Analysis contains the Prosecutor's evaluation of the incident.
	//
	// Filled by Prosecutor, empty at Officer level
	Analysis *ProsecutorAnalysis
}

// K8sEvent represents a single Kubernetes event associated with the Pod.
//
// These events are the primary source of truth for diagnosing the failure.
type K8sEvent struct {
	// Type of event (Normal, Warning, etc.)
	Type string

	// Reason is the Kubernetes event reason (e.g., Failed, Killing, BackOff)
	Reason string

	// Message is the human-readable event description
	Message string
}

// ContainerIssue represents the observed failure state of a container
// inside a Pod that has already been classified as a known incident type.
type ContainerIssue struct {
	// Container is the name of the container within the Pod.
	Container string

	// Reason is the raw Kubernetes termination reason.
	Reason string

	// Logs contains a bounded snapshot of container logs relevant to the failure.
	Logs []string
}

// ProsecutorAnalysis represents the result of analysing an incident,
// including contextual inputs and the generated commentary.
type ProsecutorAnalysis struct {
	// RelatedRepoURL points to the repository likely associated
	// with the workload incident.
	RelatedRepoURL string

	// Commentary contains the human-readable analysis generated
	// by the Prosecutor, LLM-based reasoning.
	Commentary string
}
