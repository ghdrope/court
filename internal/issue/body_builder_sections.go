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

// buildEventsSection groups incident events into warnings and normal events
// and renders them in Markdown format.
func buildEventsSection(inc *incident.IncidentReport) string {
	var normal, warnings []string

	for _, e := range inc.Events {
		line := fmt.Sprintf("- [%s] %s: %s", e.Type, e.Reason, e.Message)

		if strings.EqualFold(e.Type, "warning") {
			warnings = append(warnings, line)
		} else {
			normal = append(normal, line)
		}
	}

	if len(warnings) == 0 && len(normal) == 0 {
		return "No events available"
	}

	var b strings.Builder

	if len(warnings) > 0 {
		b.WriteString("### Warnings\n```\n")
		b.WriteString(strings.Join(warnings, "\n"))
		b.WriteString("\n```\n\n")
	}

	if len(normal) > 0 {
		b.WriteString("### Events\n```\n")
		b.WriteString(strings.Join(normal, "\n"))
		b.WriteString("\n```")
	}

	return b.String()
}

// buildContainersSection renders container-related failure data
// including image name and associated logs.
func buildContainersSection(inc *incident.IncidentReport) string {
	if len(inc.ContainersMetadata) == 0 {
		return "_No container issues detected_"
	}

	var out []string

	for _, c := range inc.ContainersMetadata {
		entry := fmt.Sprintf(
			"### Container `%s`\n- Image: `%s`\n#### Logs\n```\n%s\n```",
			c.Container,
			c.ImageName,
			formatLogs(c.Logs),
		)

		out = append(out, entry)
	}

	return strings.Join(out, "\n\n")
}
