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
	"errors"
	"testing"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/suit"
	"github.com/ghdrope/court/pkg/testhelper"
	"go.uber.org/zap"
)

// fakeSuitRepo is an in-memory implementation of SuitRepository.
// It is used to validate service behavior without requiring a real database.
type fakeSuitRepo struct {
	store     map[string]*suit.Suit
	insertErr error
}

func newFakeSuitRepo() *fakeSuitRepo {
	return &fakeSuitRepo{
		store: make(map[string]*suit.Suit),
	}
}

// Insert simulates storing a Suit in the repository.
func (f *fakeSuitRepo) Insert(ctx context.Context, s *suit.Suit) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.store[s.IncidentID] = s
	return nil
}

// GetByIncidentID simulates retrieving a Suit by incident ID.
func (f *fakeSuitRepo) GetByIncidentID(ctx context.Context, incidentID string) (*suit.Suit, error) {
	if s, ok := f.store[incidentID]; ok {
		return s, nil
	}
	return nil, nil
}

// Close is a no-op for this test implementation.
func (f *fakeSuitRepo) Close(ctx context.Context, id string) error {
	return nil
}

// TestCreateSuit_Success verifies the normal creation flow.
//
// It ensures that:
//   - a suit is created when it does not exist
//   - the repository Insert method is called
//   - no error is returned
func TestCreateSuit_Success(t *testing.T) {

	repo := newFakeSuitRepo()
	gh := &testhelper.GitHubMock{}
	logger := zap.NewNop()

	svc := New(repo, gh, logger)

	inc := &incident.IncidentReport{
		ID:            "incident-123",
		GitHubRepoURL: "https://github.com/ghdrope/court",
	}

	err := svc.CreateSuit(context.Background(), inc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, ok := repo.store["incident-123"]
	if !ok {
		t.Fatal("expected suit to be created")
	}

	if s.IncidentID != "incident-123" {
		t.Fatalf("expected incident_id incident-123, got %s", s.IncidentID)
	}

	if !gh.Called {
		t.Fatal("expected github issue to be created")
	}
}

// TestCreateSuit_EmptyIncidentID verifies input validation.
//
// It ensures that an empty incidentID returns an error.
func TestCreateSuit_EmptyIncidentID(t *testing.T) {

	repo := newFakeSuitRepo()
	gh := &testhelper.GitHubMock{}
	logger := zap.NewNop()

	svc := New(repo, gh, logger)

	err := svc.CreateSuit(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for empty incidentID")
	}
}

// TestCreateSuit_Idempotency verifies that duplicate suits are not created.
//
// It ensures that:
//   - first call creates a suit
//   - second call does not create duplicates
func TestCreateSuit_Idempotency(t *testing.T) {

	repo := newFakeSuitRepo()
	gh := &testhelper.GitHubMock{}
	logger := zap.NewNop()

	svc := New(repo, gh, logger)

	inc := &incident.IncidentReport{
		ID: "incident-456",
	}

	_ = svc.CreateSuit(context.Background(), inc)
	first := len(repo.store)

	_ = svc.CreateSuit(context.Background(), inc)
	second := len(repo.store)

	if second != first {
		t.Fatalf("expected idempotent behavior, got %d vs %d", second, first)
	}
}

// TestCreateSuit_InsertError verifies repository failure propagation.
//
// It ensures that:
//   - Insert errors are returned
//   - errors are not swallowed
func TestCreateSuit_InsertError(t *testing.T) {

	expectedErr := errors.New("insert failed")

	repo := &fakeSuitRepo{
		store:     make(map[string]*suit.Suit),
		insertErr: expectedErr,
	}

	gh := &testhelper.GitHubMock{}
	logger := zap.NewNop()

	svc := New(repo, gh, logger)

	inc := &incident.IncidentReport{
		ID: "incident-123",
	}

	err := svc.CreateSuit(context.Background(), inc)

	if err == nil {
		t.Fatal("expected error from repository")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
