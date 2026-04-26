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

package env

import (
	"testing"

	testhelper "github.com/ghdrope/court/internal/testhelper"
)

// TestRequire verifies the behavior of required environment variables.
func TestRequire(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		expectError bool
	}{
		{
			name:        "value is set",
			key:         "TEST_KEY",
			value:       "value",
			expectError: false,
		},
		{
			name:        "value is empty",
			key:         "EMPTY_KEY",
			value:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cleanup := testhelper.SetEnv(tt.key, tt.value)
			defer cleanup()

			val, err := Require(tt.key)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				if _, ok := err.(MissingError); !ok {
					t.Fatalf("expected MissingError, got %T", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if val != tt.value {
				t.Errorf("unexpected value: got %q, want %q", val, tt.value)
			}
		})
	}
}

// TestMust verifies that Must returns the value or panics when missing.
func TestMust(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		cleanup := testhelper.SetEnv("MUST_KEY", "ok")
		defer cleanup()

		val := Must("MUST_KEY")

		if val != "ok" {
			t.Errorf("unexpected value: got %q, want %q", val, "ok")
		}
	})

	t.Run("panic on missing value", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic, got none")
			}
		}()

		_ = Must("MISSING_KEY")
	})
}
