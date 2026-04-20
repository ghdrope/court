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
	"fmt"
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

	var issues []incident.ContainerIssue

	for _, cs := range pod.Status.ContainerStatuses {

		add := func(reason string) {

			issue := incident.ContainerIssue{
				Container: cs.Name,
				ImageName: cs.Image,
				Reason:    reason,
			}

			// Only fetch logs if issue is meaningful
			// (avoid overhead for healthy containers)
			logs := fetchContainerLogs(
				ctx,
				client,
				pod.Namespace,
				pod.Name,
				cs.Name,
			)

			if len(logs) == 0 {
				issue.Logs = []string{"<no logs available>"}
			} else {
				issue.Logs = logs
			}

			issues = append(issues, issue)
		}

		// Detect containers stuck in waiting states due to runtime issues.
		if cs.State.Terminated != nil {

			if cs.State.Terminated.ExitCode != 0 {
				reason := fmt.Sprintf(
					"%s (exit=%d)",
					cs.State.Terminated.Reason,
					cs.State.Terminated.ExitCode,
				)
				add(reason)
			}

			continue
		}

		if cs.State.Waiting != nil {

			switch cs.State.Waiting.Reason {

			case "CrashLoopBackOff",
				"ImagePullBackOff",
				"ErrImagePull",
				"RunContainerError":

				add(cs.State.Waiting.Reason)

			default:
				continue
			}
		}
	}

	return issues
}

func fetchContainerLogs(
	ctx context.Context,
	client kubernetes.Interface,
	namespace string,
	podName string,
	container string,
) []string {

	const maxLines = 150

	readLogs := func(previous bool) []string {
		req := client.CoreV1().
			Pods(namespace).
			GetLogs(podName, &v1.PodLogOptions{
				Container: container,
				Previous:  previous,
			})

		stream, err := req.Stream(ctx)
		if err != nil {
			return nil
		}
		defer func() {
			_ = stream.Close()
		}()

		var buffer bytes.Buffer
		_, _ = io.Copy(&buffer, stream)

		scanner := bufio.NewScanner(&buffer)

		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if len(lines) > maxLines {
			return lines[len(lines)-maxLines:]
		}

		return lines
	}

	// 1. current logs
	current := readLogs(false)

	// 2. previous logs (crucial for crashes)
	previous := readLogs(true)

	// merge with separator
	if len(previous) == 0 {
		return current
	}

	out := make([]string, 0, len(previous)+len(current)+2)

	out = append(out, "--- PREVIOUS CONTAINER LOGS ---")
	out = append(out, previous...)
	out = append(out, "--- CURRENT CONTAINER LOGS ---")
	out = append(out, current...)

	return out
}
