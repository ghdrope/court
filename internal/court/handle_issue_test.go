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
	"github.com/ghdrope/court/internal/suit"
	"go.uber.org/zap"
)

// TestCreateIssue_Success verifies that a VCS issue is created
// and attached to the Suit entity.
func TestCreateIssue_Success(t *testing.T) {
	ctx := context.Background()

	repo := NewFakeSuitRepository()
	vcsClient := &FakeVCSClient{}

	svc := New(repo, vcsClient, zap.NewNop())

	inc := &incident.IncidentReport{
		ID:         "inc-1",
		VCSRepoURL: "https://github.com/testowner/testrepo",
	}

	suitEntity := &suit.Suit{
		ID:         "suit-1",
		IncidentID: inc.ID,
	}

	err := svc.CreateIssue(ctx, inc, suitEntity)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if suitEntity.VCSIssueURL == "" {
		t.Fatal("expected VCSIssueURL to be set")
	}

	if len(vcsClient.CreatedIssues) != 1 {
		t.Fatal("expected issue to be created in VCS client")
	}
}

// TestCreateIssue_NoVCS verifies that no panic occurs when VCS is nil.
func TestCreateIssue_NoVCS(t *testing.T) {
	ctx := context.Background()

	repo := NewFakeSuitRepository()

	svc := New(repo, nil, zap.NewNop())

	inc := &incident.IncidentReport{
		ID: "inc-1",
	}

	suitEntity := &suit.Suit{ID: "suit-1"}

	err := svc.CreateIssue(ctx, inc, suitEntity)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
