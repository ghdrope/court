package inspector

import (
	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
)

// DetectContainerIssues analyses the container statuses of a Pod
// and returns a list of detected issues, such as CrashLoopBackOff,
// ImagePullBackOff, or OOMKilled conditions.
//
// Does not perform logging and is intended solely as an inspection utility.
func DetectContainerIssues(pod *v1.Pod) []incident.ContainerIssue {

	if pod == nil {
		return nil
	}

	issues := make([]incident.ContainerIssue, 0, len(pod.Status.ContainerStatuses))

	for _, cs := range pod.Status.ContainerStatuses {

		add := func(reason string) {
			issues = append(issues, incident.ContainerIssue{
				Container: cs.Name,
				Reason:    reason,
			})
		}

		// Detect containers stuck in waiting states due to runtime issues.
		if cs.State.Waiting != nil {
			switch reason := cs.State.Waiting.Reason; reason {
			case "CrashLoopBackOff", "ImagePullBackOff":
				add(reason)
			}
		}

		// Detect containers terminated due to out-of-memory conditions.
		if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
			add(cs.State.Terminated.Reason)
		}
	}

	return issues
}
