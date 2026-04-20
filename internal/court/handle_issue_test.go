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
	"testing"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/pkg/testhelper"
	"go.uber.org/zap"
)

// TestCreateGitHubIssue verifies that a GitHub issue is correctly created
// from an IncidentReport.
func TestCreateGitHubIssue(t *testing.T) {

	gh := &testhelper.GitHubMock{}

	svc := &Service{
		GitHub: gh,
		Log:    zap.NewNop(),
	}

	inc := &incident.IncidentReport{
		ID:            "incident-123",
		Cluster:       "test-cluster",
		Namespace:     "default",
		Pod:           "api-pod",
		GitHubRepoURL: "https://github.com/ghdrope/court",

		Events: []incident.K8sEvent{
			{
				Type:    "Warning",
				Reason:  "CrashLoopBackOff",
				Message: "container crashed repeatedly",
			},
		},

		ContainerIssues: []incident.ContainerIssue{
			{
				Container: "api",
				Reason:    "CrashLoopBackOff",
				Logs:      []string{"error line 1", "error line 2"},
			},
		},
	}

	url, err := svc.createGitHubIssue(context.Background(), inc)

	// Assertions: no error expected
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Must return a URL from mock
	if url == "" {
		t.Fatal("expected issue URL, got empty string")
	}

	// Ensure GitHub client was invoked
	if !gh.Called {
		t.Fatal("expected GitHub CreateIssue to be called")
	}

	// Validate title composition
	if gh.Title != "Court Incident incident-123" {
		t.Errorf("unexpected title: %s", gh.Title)
	}

	// Validate title composition
	if gh.Title != "Court Incident incident-123" {
		t.Errorf("unexpected title: got %s", gh.Title)
	}

	// Validate body generation
	if gh.Body == "" {
		t.Fatal("expected non-empty issue body")
	}

	if len(gh.Body) < 50 {
		t.Error("issue body seems too short")
	}
}

// TestBuildIssueBody verifies that the issue body builder produces
// a meaningful and structured output.
func TestBuildIssueBody(t *testing.T) {

	inc := &incident.IncidentReport{
		ID:        "inc-1",
		Cluster:   "c1",
		Namespace: "ns1",
		Pod:       "pod1",
		Events: []incident.K8sEvent{
			{
				Type:    "Normal",
				Reason:  "Started",
				Message: "container started",
			},
		},
		ContainerIssues: []incident.ContainerIssue{
			{
				Container: "app",
				Reason:    "Error",
				Logs:      []string{"log1", "log2", "log3", "log4"},
			},
		},
	}

	body := buildIssueBody(inc)

	if body == "" {
		t.Fatal("expected body not to be empty")
	}

	if !testhelper.Contains(body, "Court Incident") {
		t.Error("missing title section")
	}

	if !testhelper.Contains(body, "Container: app") {
		t.Error("missing container section")
	}

	if !testhelper.Contains(body, "log1") {
		t.Error("missing logs")
	}
}
