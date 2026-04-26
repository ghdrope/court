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

	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/ghdrope/court/internal/suit"
)

// fakeSuitRepo implements SuitRepository using real domain type.
type fakeSuitRepo struct {
	suits []suit.Suit
	err   error
}

func (f *fakeSuitRepo) ListOpen(ctx context.Context) ([]suit.Suit, error) {
	return f.suits, f.err
}

// TestRecoverOpenIncidents verifies recovery executes safely when no suits exist
// and no recovery hints are generated.
func TestRecoverOpenIncidents(t *testing.T) {

	tests := []struct {
		name        string
		suits       []suit.Suit
		expectHints int
	}{
		{
			name:        "empty suits",
			suits:       []suit.Suit{},
			expectHints: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc := &Service{
				Log:      logr.Discard(),
				RDB:      nil,
				SuitRepo: &fakeSuitRepo{suits: tt.suits},
			}

			kube := fake.NewSimpleClientset()

			hints, err := svc.RecoverOpenIncidents(
				context.TODO(),
				"cluster-1",
				kube,
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(hints) != tt.expectHints {
				t.Errorf("expected %d hints, got %d", tt.expectHints, len(hints))
			}
		})
	}
}
