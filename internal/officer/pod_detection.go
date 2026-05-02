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
	"context"
	"fmt"
	"time"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// minPodAge defines a grace period before evaluating a Pod.
	// This avoids false positives during initial scheduling/startup.
	minPodAge = 30 * time.Second

	// maxLogBytes limits the amount of logs fetched per container.
	// This prevents excessive memory usage and large payloads.
	maxLogBytes = 2 * 1024 * 1024
)

// DetectContainerIssues inspects container statuses and extracts failure signals.
//
// It detects:
//   - terminated containers with non-zero exit codes
//   - containers in known failure waiting states
//
// Pods younger than minPodAge are ignored to avoid startup noise.
func DetectContainersMetadata(
	ctx context.Context,
	client kubernetes.Interface,
	pod *v1.Pod,
) []incident.ContainerMetadata {

	if pod == nil {
		return nil
	}

	if time.Since(pod.CreationTimestamp.Time) < minPodAge {
		return nil
	}

	var issues []incident.ContainerMetadata

	for _, cs := range pod.Status.ContainerStatuses {

		// Terminated container failures
		if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
			issues = append(issues, incident.ContainerMetadata{
				Container: cs.Name,
				ImageName: cs.Image,
				Reason: fmt.Sprintf(
					"terminated (%s exit=%d)",
					cs.State.Terminated.Reason,
					cs.State.Terminated.ExitCode,
				),
			})
			continue
		}

		// Waiting state failures
		if cs.State.Waiting != nil {
			switch cs.State.Waiting.Reason {
			case "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull", "RunContainerError":
				issues = append(issues, incident.ContainerMetadata{
					Container: cs.Name,
					ImageName: cs.Image,
					Reason:    cs.State.Waiting.Reason,
				})
			}
		}
	}

	return issues
}
