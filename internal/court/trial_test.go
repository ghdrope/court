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
	"go.uber.org/zap"
)

// TestCreateTrial_EndToEnd verifies full lifecycle:
// Suit creation + issue creation + persistence.
func TestCreateTrial_EndToEnd(t *testing.T) {
	ctx := context.Background()

	repo := NewFakeSuitRepository()
	vcs := &FakeVCSClient{}

	svc := New(repo, vcs, zap.NewNop())

	inc := &incident.IncidentReport{
		ID:         "inc-1",
		VCSRepoURL: "https://github.com/test/test",
	}

	err := svc.CreateTrial(ctx, inc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	suitEntity, _ := repo.GetByIncidentID(ctx, inc.ID)

	if suitEntity == nil {
		t.Fatal("expected suit to exist")
	}

	if suitEntity.VCSIssueURL == "" {
		t.Fatal("expected issue URL to be set")
	}
}

// TestCloseTrial_NoSuit ensures graceful no-op when no suit exists.
func TestCloseTrial_NoSuit(t *testing.T) {
	ctx := context.Background()

	repo := NewFakeSuitRepository()
	svc := New(repo, &FakeVCSClient{}, zap.NewNop())

	err := svc.CloseTrial(ctx, "missing", "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
