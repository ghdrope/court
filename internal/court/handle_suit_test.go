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

	"github.com/ghdrope/court/internal/suit"
	"go.uber.org/zap"
)

// fakeSuitRepo is a minimal in-memory implementation of the suit repository.
//
// It is used to validate the behavior of CreateSuit without requiring a real database.
type fakeSuitRepo struct {
	store     map[string]*suit.Suit
	insertErr error
}

// newFakeSuitRepo initializes an empty in-memory repository.
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
	logger := zap.NewNop()

	svc := New(repo, logger)

	err := svc.CreateSuit(context.Background(), "incident-123")
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
}

// TestCreateSuit_EmptyIncidentID verifies input validation.
//
// It ensures that an empty incidentID returns an error.
func TestCreateSuit_EmptyIncidentID(t *testing.T) {

	repo := newFakeSuitRepo()
	logger := zap.NewNop()

	svc := New(repo, logger)

	err := svc.CreateSuit(context.Background(), "")
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
	logger := zap.NewNop()

	svc := New(repo, logger)

	incidentID := "incident-456"

	_ = svc.CreateSuit(context.Background(), incidentID)
	firstCount := len(repo.store)

	_ = svc.CreateSuit(context.Background(), incidentID)
	secondCount := len(repo.store)

	if firstCount != 1 {
		t.Fatalf("expected 1 suit after first call, got %d", firstCount)
	}

	if secondCount != 1 {
		t.Fatalf("expected idempotent behavior, got %d suits", secondCount)
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

	logger := zap.NewNop()

	svc := New(repo, logger)

	err := svc.CreateSuit(context.Background(), "incident-123")

	if err == nil {
		t.Fatal("expected error from repository")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
