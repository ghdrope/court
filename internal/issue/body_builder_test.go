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
	"testing"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/testhelper"
)

// TestBuildEventsSection_NoEvents ensures empty event lists return fallback message.
func TestBuildEventsSection_NoEvents(t *testing.T) {
	t.Parallel()

	inc := &incident.IncidentReport{
		Events: []incident.K8sEvent{},
	}

	got := buildEventsSection(inc)

	if got != "No events available" {
		t.Errorf("expected fallback message, got %s", got)
	}
}

// TestBuildEventsSection_WarningsAndEvents ensures warnings and normal events
// are properly separated and rendered.
func TestBuildEventsSection_WarningsAndEvents(t *testing.T) {
	t.Parallel()

	inc := &incident.IncidentReport{
		Events: []incident.K8sEvent{
			{Type: "warning", Reason: "CPU", Message: "high usage"},
			{Type: "info", Reason: "startup", Message: "container started"},
		},
	}

	got := buildEventsSection(inc)

	if got == "" {
		t.Fatalf("expected non-empty result")
	}

	if got != "No events available" &&
		(!testhelper.Contains(got, "Warnings") || !testhelper.Contains(got, "Events")) {
		t.Errorf("expected warnings and events sections, got: %s", got)
	}
}

// TestBuildContainersSection_NoContainers ensures empty container list
// returns fallback message.
func TestBuildContainersSection_NoContainers(t *testing.T) {
	t.Parallel()

	inc := &incident.IncidentReport{
		ContainersMetadata: []incident.ContainerMetadata{},
	}

	got := buildContainersSection(inc)

	if got != "_No container issues detected_" {
		t.Errorf("unexpected result: %s", got)
	}
}

// TestBuildContainersSection_WithContainers ensures container issues
// are rendered with image and logs.
func TestBuildContainersSection_WithContainers(t *testing.T) {
	t.Parallel()

	inc := &incident.IncidentReport{
		ContainersMetadata: []incident.ContainerMetadata{
			{
				Container: "api",
				ImageName: "nginx:latest",
				Logs:      []string{"error starting server"},
			},
		},
	}

	got := buildContainersSection(inc)

	if got == "" {
		t.Fatalf("expected non-empty result")
	}

	if !testhelper.Contains(got, "Container `api`") {
		t.Errorf("expected container section, got: %s", got)
	}

	if !testhelper.Contains(got, "nginx:latest") {
		t.Errorf("expected image name in output")
	}
}
