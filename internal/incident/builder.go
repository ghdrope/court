package incident

import (
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
)

// BuildFromPod constructs an IncidentReport from a given Pod,
// extracting its metadata, phase, and detected container issues.
// Logs are provided externally and attached as-is.
//
// This function acts as a translation layer between raw Kubernetes
// Pod state and the domain-level IncidentReport model.
func BuildFromPod(pod *v1.Pod, containerIssues []ContainerIssue, logs []string) IncidentReport {

	reason := string(pod.Status.Phase)

	switch pod.Status.Phase {
	case v1.PodFailed:
		reason = "PodFailed"
	case v1.PodUnknown:
		reason = "PodUnknown"
	}

	return IncidentReport{
		EventID: uuid.NewString(),

		PodName:   pod.Name,
		Namespace: pod.Namespace,
		Phase:     pod.Status.Phase,
		Reason:    reason,

		ContainerIssues: containerIssues,
		Logs:            logs,
	}
}
