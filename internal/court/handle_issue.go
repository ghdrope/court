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
	"net/url"
	"strings"

	"github.com/ghdrope/court/internal/incident"
)

// createGitHubIssue builds and sends the issue to GitHub.
//
// It uses the IncidentReport GitHubRepoURL to determine the target repository.
// If the repository cannot be resolved, issue creation is skipped.
func (s *Service) createGitHubIssue(ctx context.Context, inc *incident.IncidentReport) (string, error) {

	if s.GitHub == nil {
		return "", nil
	}

	if inc.GitHubRepoURL == "" {
		return "", fmt.Errorf("missing GitHubRepoURL on incident")
	}

	owner, repo, err := parseGitHubRepo(inc.GitHubRepoURL)
	if err != nil {
		return "", fmt.Errorf("invalid repository url: %w", err)
	}

	title := fmt.Sprintf("Court Incident %s", inc.ID)
	body := buildIssueBody(inc)

	return s.GitHub.CreateIssue(ctx, owner, repo, title, body)
}

// parseGitHubRepo extracts owner/repo from a GitHub URL.
//
// Example:
//
//	https://github.com/org/repo -> org, repo
func parseGitHubRepo(raw string) (string, string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid github repo path")
	}

	return parts[0], parts[1], nil
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
			max := 5
			truncated := false

			if len(i.Logs) > max {
				truncated = true
			}

			limit := max
			if len(i.Logs) < max {
				limit = len(i.Logs)
			}

			logs = strings.Join(i.Logs[:limit], "\n")

			if truncated {
				logs += "\n... (truncated)"
			}
		}

		entry := fmt.Sprintf(
			"- Container: `%s`\n"+
				"  Image: `%s`\n"+
				"  Reason: `%s`\n"+
				"  Logs:\n\n"+
				"```bash\n%s\n```",
			i.Container,
			i.ImageName,
			i.Reason,
			formatLogs(logs),
		)

		issues = append(issues, entry)
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

---

💡 As a productivity tip you can use GitHub Copilot to help investigate and fix this issue.

Thank you for using Court!
`,
		inc.Cluster,
		inc.Namespace,
		inc.Pod,
		joinOrFallback(events),
		joinOrFallback(issues),
	)
}

// joinOrFallback ensures empty sections are still readable.
func joinOrFallback(items []string) string {
	if len(items) == 0 {
		return "_No data available_"
	}
	return strings.Join(items, "\n")
}

// indentLogs ensures logs render correctly inside GitHub Markdown list items.
// Without indentation, fenced code blocks break out of list context.
func formatLogs(logs string) string {
	// ensures logs are safe inside fenced code block
	logs = strings.TrimSpace(logs)

	if logs == "" {
		return "<no logs available>"
	}

	return logs
}
