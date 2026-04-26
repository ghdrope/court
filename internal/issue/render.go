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

package issue

import (
	"fmt"
	"strings"

	"github.com/ghdrope/court/internal/incident"
)

// renderSection formats a section with a title and items.
func renderSection(title string, items []string) string {
	if len(items) == 0 {
		return ""
	}

	return fmt.Sprintf("### %s\n%s", title, strings.Join(items, "\n"))
}

// joinSections merges multiple sections safely.
func joinSections(sections ...string) string {
	var out []string

	for _, s := range sections {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}

	if len(out) == 0 {
		return "_No events available_"
	}

	return strings.Join(out, "\n\n")
}

// formatLogs ensures logs are safe for markdown rendering.
func formatLogs(logs []string) string {
	if len(logs) == 0 {
		return "<no logs available>"
	}

	return strings.Join(logs, "\n")
}

// extractPrimaryReason returns the most relevant failure reason.
func extractPrimaryReason(inc *incident.IncidentReport) string {
	if len(inc.ContainerIssues) > 0 {
		if inc.ContainerIssues[0].Reason != "" {
			return inc.ContainerIssues[0].Reason
		}
	}
	return "unknown"
}
