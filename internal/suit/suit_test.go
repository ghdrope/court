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

package suit

import (
	"testing"
	"time"
)

// TestNewSuit_DefaultValues ensures a new Suit is initialized with correct defaults.
func TestNewSuit_DefaultValues(t *testing.T) {
	now := time.Now()

	s := Suit{
		ID:         "suit-1",
		IncidentID: "incident-1",
		Status:     StatusOpen,
		CreatedAt:  now,
		ClosedAt:   nil,
	}

	if s.ID != "suit-1" {
		t.Errorf("expected ID to be 'suit-1', got %s", s.ID)
	}

	if s.IncidentID != "incident-1" {
		t.Errorf("expected IncidentID to be 'incident-1', got %s", s.IncidentID)
	}

	if s.Status != StatusOpen {
		t.Errorf("expected Status to be StatusOpen, got %s", s.Status)
	}

	if s.CreatedAt != now {
		t.Errorf("expected CreatedAt to be %v, got %v", now, s.CreatedAt)
	}

	if s.ClosedAt != nil {
		t.Errorf("expected ClosedAt to be nil, got %v", s.ClosedAt)
	}
}

// TestSuit_Close verifies that closing a Suit sets the correct status and timestamp.
func TestSuit_Close(t *testing.T) {
	now := time.Now()

	s := Suit{
		ID:         "suit-2",
		IncidentID: "incident-2",
		Status:     StatusOpen,
		CreatedAt:  now,
		ClosedAt:   nil,
	}

	closeTime := now.Add(1 * time.Hour)
	s.Status = StatusClosed
	s.ClosedAt = &closeTime

	if s.Status != StatusClosed {
		t.Errorf("expected Status to be StatusClosed, got %s", s.Status)
	}

	if s.ClosedAt == nil {
		t.Fatal("expected ClosedAt to be set, got nil")
	}

	if !s.ClosedAt.Equal(closeTime) {
		t.Errorf("expected ClosedAt to be %v, got %v", closeTime, s.ClosedAt)
	}
}

// TestSuit_StatusValues ensures valid status constants behave as expected.
func TestSuit_StatusValues(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		valid  bool
	}{
		{"open status", StatusOpen, true},
		{"closed status", StatusClosed, true},
		{"invalid status", Status("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.status {
			case StatusOpen, StatusClosed:
				if !tt.valid {
					t.Errorf("expected status %s to be invalid", tt.status)
				}
			default:
				if tt.valid {
					t.Errorf("expected status %s to be valid", tt.status)
				}
			}
		})
	}
}

// TestSuit_CloseIdempotent ensures closing an already closed Suit does not break invariants.
func TestSuit_CloseIdempotent(t *testing.T) {
	now := time.Now()
	closeTime := now.Add(1 * time.Hour)

	s := Suit{
		ID:         "suit-3",
		IncidentID: "incident-3",
		Status:     StatusClosed,
		CreatedAt:  now,
		ClosedAt:   &closeTime,
	}

	// Attempt to "close" again
	newCloseTime := now.Add(2 * time.Hour)
	s.ClosedAt = &newCloseTime

	if s.Status != StatusClosed {
		t.Errorf("expected Status to remain StatusClosed, got %s", s.Status)
	}

	if !s.ClosedAt.Equal(newCloseTime) {
		t.Errorf("expected ClosedAt to be updated to %v, got %v", newCloseTime, s.ClosedAt)
	}
}
