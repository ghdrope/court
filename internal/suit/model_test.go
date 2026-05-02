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
)

// TestStatusValues ensures that Status constants are correctly defined.
func TestStatusValues(t *testing.T) {
	t.Parallel()

	if StatusOpen != "open" {
		t.Errorf("expected StatusOpen to be 'open', got %s", StatusOpen)
	}

	if StatusClosed != "closed" {
		t.Errorf("expected StatusClosed to be 'closed', got %s", StatusClosed)
	}
}

// TestSuitStructValidation ensures Suit struct fields behave as expected.
func TestSuitStructValidation(t *testing.T) {
	t.Parallel()

	s := &Suit{
		Status: StatusOpen,
	}

	if s.Status != StatusOpen {
		t.Errorf("expected status open, got %s", s.Status)
	}

	if s.ClosedAt != nil {
		t.Errorf("expected ClosedAt to be nil for open suit")
	}
}
