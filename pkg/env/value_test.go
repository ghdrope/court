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

	testhelper "github.com/ghdrope/court/pkg/testhelper"
)

// TestGet verifies environment variable fallback behavior.
func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "env value exists",
			key:          "KEY1",
			value:        "env-value",
			defaultValue: "default",
			expected:     "env-value",
		},
		{
			name:         "env value missing",
			key:          "KEY2",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cleanup := testhelper.SetEnv(tt.key, tt.value)
			defer cleanup()

			got := Get(tt.key, tt.defaultValue)

			if got != tt.expected {
				t.Errorf("unexpected value: got %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestFirstNonEmpty verifies precedence resolution of values.
func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "first value is non-empty",
			values:   []string{"a", "b", "c"},
			expected: "a",
		},
		{
			name:     "skip empty values",
			values:   []string{"", "", "c"},
			expected: "c",
		},
		{
			name:     "all empty",
			values:   []string{"", "", ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := FirstNonEmpty(tt.values...)

			if got != tt.expected {
				t.Errorf("unexpected value: got %q, want %q", got, tt.expected)
			}
		})
	}
}
