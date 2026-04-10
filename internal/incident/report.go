package incident

import v1 "k8s.io/api/core/v1"

// IncidentReport represents a structured description of a pod failure.
// capturing essential metadata and runtime context for downstream processing.
type IncidentReport struct {
	PodName   string
	Namespace string
	Phase     v1.PodPhase
	Reason    string

	// ContainerIssues contains detected issues at the container level,
	// such as CrashLoopBackOff or OOMKilled states.
	ContainerIssues []ContainerIssue

	// Logs contains a limited snapshot of relevant pod logs.
	// This is intentionally bounded to avoid large payloads.
	Logs []string
}

// ContainerIssue represents a failure or abnormal state detected
// in a specific container within a pod.
type ContainerIssue struct {
	Container string
	Reason    string
}
