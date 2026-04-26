package officer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	minPodAge   = 30 * time.Second
	maxLogBytes = 2 * 1024 * 1024
)

func DetectContainerIssues(
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

		// TERMINATED FAILURES
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

		// WAITING FAILURES
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

func fetchContainerLogs(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, podName, container string,
) []string {

	if namespace == "court" {
		return []string{"<blocked system namespace>"}
	}

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
