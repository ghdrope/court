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

// TestCreateSuit_New verifies a new Suit is created.
func TestCreateSuit_New(t *testing.T) {
	ctx := context.Background()

	repo := NewFakeSuitRepository()
	svc := New(repo, &FakeVCSClient{}, zap.NewNop())

	inc := &incident.IncidentReport{
		ID: "inc-1",
	}

	s, err := svc.CreateSuit(ctx, inc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s == nil {
		t.Fatal("expected suit to be created")
	}

	if s.IncidentID != inc.ID {
		t.Fatal("incident ID mismatch")
	}
}

// TestCreateSuit_Existing verifies that existing Suit is reused.
func TestCreateSuit_Existing(t *testing.T) {
	ctx := context.Background()

	repo := NewFakeSuitRepository()
	svc := New(repo, &FakeVCSClient{}, zap.NewNop())

	inc := &incident.IncidentReport{ID: "inc-1"}

	existing := &suit.Suit{
		ID:         "suit-1",
		IncidentID: inc.ID,
	}

	_ = repo.Insert(ctx, existing)

	s, err := svc.CreateSuit(ctx, inc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.ID != "suit-1" {
		t.Fatal("expected existing suit to be returned")
	}
}
