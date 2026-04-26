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

package postgres

import "testing"

// TestSanitizeDSN verifies that credentials are redacted.
func TestSanitizeDSN(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid dsn",
			input:    "postgres://user:pass@localhost:5432/db",
			expected: "postgres://***:***@localhost:5432/db",
		},
		{
			name:     "no credentials",
			input:    "postgres://localhost:5432/db",
			expected: "postgres://localhost:5432/db",
		},
		{
			name:     "missing markers",
			input:    "invalid-string",
			expected: "invalid-string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeDSN(tt.input)

			if got != tt.expected {
				t.Errorf("unexpected result: got %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestRedactBetween verifies internal redaction logic.
func TestRedactBetween(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    string
		end      string
		expected string
	}{
		{
			name:     "normal case",
			input:    "a://secret@b",
			start:    "://",
			end:      "@",
			expected: "a://***:***@b",
		},
		{
			name:     "missing start",
			input:    "abc",
			start:    "://",
			end:      "@",
			expected: "abc",
		},
		{
			name:     "missing end",
			input:    "a://secret",
			start:    "://",
			end:      "@",
			expected: "a://secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactBetween(tt.input, tt.start, tt.end)

			if got != tt.expected {
				t.Errorf("unexpected result: got %q, want %q", got, tt.expected)
			}
		})
	}
}
