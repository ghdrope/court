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

package court

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghdrope/court/internal/incident"
)

// createGitHubIssue builds and sends the issue to GitHub.
func (s *Service) createGitHubIssue(ctx context.Context, inc *incident.IncidentReport) error {

	title := fmt.Sprintf("Court Incident %s", inc.ID)
	body := buildIssueBody(inc)

	return s.GitHub.CreateIssue(ctx, title, body)
}

// buildIssueBody formats the GitHub issue content.
//
// Keeps formatting logic isolated for maintainability.
func buildIssueBody(inc *incident.IncidentReport) string {

	var events []string
	for _, e := range inc.Events {
		events = append(events, fmt.Sprintf(
			"- [%s] %s: %s",
			e.Type,
			e.Reason,
			e.Message,
		))
	}

	var issues []string
	for _, i := range inc.ContainerIssues {

		logs := "no logs available"
		if len(i.Logs) > 0 {
			// Limit logs to avoid exceeding GitHub issue size limits
			max := 3
			if len(i.Logs) < max {
				max = len(i.Logs)
			}
			logs = strings.Join(i.Logs[:max], "\n")
		}

		entry := fmt.Sprintf(
			`- Container: %s
  Reason: %s
  Logs:
  `+"```"+`
%s
  `+"```",
			i.Container,
			i.Reason,
			logs,
		)

		issues = append(issues, entry)
	}

	analysis := "No analysis available."
	if inc.Analysis != nil && inc.Analysis.Commentary != "" {
		analysis = inc.Analysis.Commentary
	}

	return fmt.Sprintf(`
This is an issue auto-generated through Court.

Your incident was caught at:

- Cluster: %s
- Namespace: %s
- Pod: %s

## Events
%s

## Container Issues
%s

## Prosecutor Insights
%s

---

💡 As a productivity tip you can use GitHub Copilot to help investigate and fix this issue.

Thank you for using Court!
`,
		inc.Cluster,
		inc.Namespace,
		inc.Pod,
		joinOrFallback(events),
		joinOrFallback(issues),
		analysis,
	)
}

// joinOrFallback ensures empty sections are still readable.
func joinOrFallback(items []string) string {
	if len(items) == 0 {
		return "_No data available_"
	}
	return strings.Join(items, "\n")
}
