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

	var issues []incident.ContainerIssue

	for _, cs := range pod.Status.ContainerStatuses {

		// Detect containers stuck in waiting states due to runtime issues.
		if cs.State.Waiting != nil {
			switch cs.State.Waiting.Reason {
			case "CrashLoopBackOff", "ImagePullBackOff":
				issues = append(issues, incident.ContainerIssue{
					Container: cs.Name,
					Reason:    cs.State.Waiting.Reason,
				})
			}
		}

		// Detect containers terminated due to out-of-memory conditions.
		if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
			issues = append(issues, incident.ContainerIssue{
				Container: cs.Name,
				Reason:    "OOMKilled",
			})
		}
	}

	return issues
}
