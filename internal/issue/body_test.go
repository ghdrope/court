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

// TestBuildBody_NilIncident ensures nil input returns safe fallback body.
func TestBuildBody_NilIncident(t *testing.T) {
	t.Parallel()

	got := buildBody(nil)

	if got != "## 🚨 Incident report is nil" {
		t.Errorf("unexpected nil handling output")
	}
}

// TestBuildBody_Complete ensures full incident report is rendered
// with all major sections included.
func TestBuildBody_Complete(t *testing.T) {
	t.Parallel()

	inc := &incident.IncidentReport{
		Cluster:   "c1",
		Namespace: "ns1",
		Pod:       "api",
		Events: []incident.K8sEvent{
			{Type: "info", Reason: "start", Message: "ok"},
		},
		ContainersMetadata: []incident.ContainerMetadata{
			{
				Container: "api",
				ImageName: "nginx",
				Logs:      []string{"boot ok"},
			},
		},
		VCSRepoURL: "https://github.com/example/repo",
	}

	got := buildBody(inc)

	if got == "" {
		t.Fatalf("expected non-empty body")
	}

	if !testhelper.Contains(got, "Automated Incident Report") {
		t.Errorf("missing report header")
	}

	if !testhelper.Contains(got, "Cluster: c1") {
		t.Errorf("missing environment section")
	}
}
