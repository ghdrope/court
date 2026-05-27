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

package utils

import (
	"os"
	"testing"
)

// TestIsDebug validates the behavior of IsDebug under different
// environment variable configurations.
func TestIsDebug(t *testing.T) {

	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "debug enabled when DEBUG=true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "debug disabled when DEBUG=false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "debug disabled when DEBUG is empty",
			envValue: "",
			expected: false,
		},
		{
			name:     "debug disabled when DEUBG has invalid value",
			envValue: "yes",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			original := os.Getenv("DEBUG")

			t.Cleanup(func() {
				_ = os.Setenv("DEBUG", original)
			})

			// Set test-specific environment value.
			_ = os.Setenv("DEBUG", tt.envValue)

			got := IsDebug()

			if got != tt.expected {
				t.Errorf(
					"IsDebug() = %v, want %v (DEBUG=%q)",
					got,
					tt.expected,
					tt.envValue,
				)
			}
		})
	}
}
