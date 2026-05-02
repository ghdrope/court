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
	"log"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// EnrichContainersMetadataWithLogs attaches logs to container metadata.
//
// For each container in the Pod:
//   - ensures a metadata entry exists (even if no issue was detected)
//   - fetches logs (current and previous)
//   - assigns logs or a fallback message if unavailable
//
// This guarantees a complete evidence set for incident analysis.
func EnrichContainersMetadataWithLogs(
	ctx context.Context,
	client kubernetes.Interface,
	pod *v1.Pod,
	issues []incident.ContainerMetadata,
) []incident.ContainerMetadata {

	issueMap := map[string]*incident.ContainerMetadata{}
	for i := range issues {
		issueMap[issues[i].Container] = &issues[i]
	}

	for _, c := range pod.Spec.Containers {

		issue, exists := issueMap[c.Name]

		if !exists {
			issues = append(issues, incident.ContainerMetadata{
				Container: c.Name,
				ImageName: c.Image,
				Reason:    "no detected failure (log enrichment)",
			})
			issue = &issues[len(issues)-1]
		}

		logs := fetchContainerLogs(
			ctx,
			client,
			pod.Namespace,
			pod.Name,
			c.Name,
		)

		if len(logs) == 0 {
			issue.Logs = []string{"<no logs available>"}
		} else {
			issue.Logs = logs
		}
	}

	return issues
}

// fetchContainerLogs retrieves logs for a specific container.
//   - fetches both current and previous logs (if available)
//   - applies a size limit (maxLogBytes)
//   - returns logs as individual lines
//
// Special cases:
//   - system namespace ("court") is blocked
//   - errors are returned as synthetic log lines
func fetchContainerLogs(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, podName, container string,
) []string {

	if client == nil {
		return []string{"<no client available>"}
	}

	if namespace == "court" {
		return []string{"<blocked system namespace>"}
	}

	// readLogs fetches logs for either current or previous container instance.
	readLogs := func(previous bool) []string {

		req := client.CoreV1().
			Pods(namespace).
			GetLogs(podName, &v1.PodLogOptions{
				Container: container,
				Previous:  previous,
			})

		stream, err := req.Stream(ctx)
		if err != nil {
			return []string{fmt.Sprintf("<log error: %v>", err)}
		}
		defer func() {
			if err := stream.Close(); err != nil {
				log.Printf("failed to close stream: %v", err)
			}
		}()

		// readLogs fetches logs for either current or previous container instance.
		limited := io.LimitReader(stream, maxLogBytes)

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, limited)

		scanner := bufio.NewScanner(&buf)

		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		return lines
	}

	current := readLogs(false)
	previous := readLogs(true)

	// If no previous logs, return current only
	if len(previous) == 0 {
		return current
	}

	out := make([]string, 0, len(previous)+len(current)+2)

	if len(previous) > 0 {
		out = append(out, "--- previous logs ---")
		out = append(out, previous...)
	}

	if len(current) > 0 {
		out = append(out, "--- current logs ---")
		out = append(out, current...)
	}

	return out
}
