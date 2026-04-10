package inspector

import (
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
)

// DetectContainerIssues inspects container statuses for real runtime failures
func DetectContainerIssues(log logr.Logger, pod *v1.Pod) {

	for _, cs := range pod.Status.ContainerStatuses {

		// Waiting states
		if cs.State.Waiting != nil {
			switch cs.State.Waiting.Reason {
			case "CrashLoopBackOff", "ImagePullBackOff":
				log.Info("container issue detected",
					"container", cs.Name,
					"reason", cs.State.Waiting.Reason,
				)
			}
		}

		// Container terminated abnormally
		if cs.State.Terminated != nil {
			if cs.State.Terminated.Reason == "OOMKilled" {
				log.Info("container killed due to OOM",
					"container", cs.Name,
				)
			}
		}
	}
}
