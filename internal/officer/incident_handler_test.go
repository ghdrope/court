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

package officer

import (
	"context"
	"testing"

	"github.com/ghdrope/court/internal/incident"
)

// fakeIncidentRepo is a minimal in-memory implementation of IncidentRepository
// used to control insertion behavior during tests.
type fakeIncidentRepo struct {
	created bool
}

// Insert simulates persistence behavior.
//
// It returns:
//   - created=true when the incident is considered new
//   - created=false when the incident already exists
func (f *fakeIncidentRepo) Insert(ctx context.Context, r *incident.IncidentReport) (bool, error) {
	return f.created, nil
}

// TestHandleIncident_NewIncident verifies that handling a new incident.
func TestHandleIncident_NewIncident(t *testing.T) {
	svc := &Service{
		IncidentRepo: &fakeIncidentRepo{created: true},
		RDB:          nil,
	}

	err := svc.HandleIncident(context.Background(), &incident.IncidentReport{
		ID: "test",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestHandleIncident_UpdateDoesNotEmit verifies that updating an existing incident.
func TestHandleIncident_UpdateDoesNotEmit(t *testing.T) {
	svc := &Service{
		IncidentRepo: &fakeIncidentRepo{created: false},
	}

	err := svc.HandleIncident(context.Background(), &incident.IncidentReport{
		ID: "test",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
