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
	"bufio"
	"bytes"
	"context"
	"io"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// DetectContainerIssues analyses the container statuses of a Pod
// and returns a list of detected issues such as:
//
// - CrashLoopBackOff,
// - ImagePullBackOff
// - OOMKilled
func DetectContainerIssues(ctx context.Context, client kubernetes.Interface, pod *v1.Pod) []incident.ContainerIssue {

	if pod == nil {
		return nil
	}

	issues := make([]incident.ContainerIssue, 0, len(pod.Status.ContainerStatuses))

	for _, cs := range pod.Status.ContainerStatuses {

		add := func(reason string) {

			issue := incident.ContainerIssue{
				Container: cs.Name,
				Reason:    reason,
			}

			// Only fetch logs if issue is meaningful
			// (avoid overhead for healthy containers)
			if pod.Namespace != "" {
				issue.Logs = fetchContainerLogs(
					ctx,
					client,
					pod.Namespace,
					pod.Name,
					cs.Name,
				)
			}

			issues = append(issues, issue)
		}

		// Detect containers stuck in waiting states due to runtime issues.
		if cs.State.Waiting != nil {
			switch cs.State.Waiting.Reason {
			case "CrashLoopBackOff", "ImagePullBackOff":
				add(cs.State.Waiting.Reason)
			}
		}

		// Detect containers terminated due to out-of-memory conditions.
		if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
			add(cs.State.Terminated.Reason)
		}
	}

	return issues
}

// fetchContainerLogs retrieves the last 200 log lines for a container.
func fetchContainerLogs(
	ctx context.Context,
	client kubernetes.Interface,
	namespace string,
	pod string,
	container string,
) []string {

	const maxLines = 200

	req := client.CoreV1().
		Pods(namespace).
		GetLogs(pod, &v1.PodLogOptions{
			Container: container,
		})

	stream, err := req.Stream(ctx)
	if err != nil {
		return nil
	}
	defer stream.Close()

	var buffer bytes.Buffer
	_, _ = io.Copy(&buffer, stream)

	scanner := bufio.NewScanner(&buffer)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) <= maxLines {
		return lines
	}

	return lines[len(lines)-maxLines:]
}
