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
	"os"
	"testing"
)

// TestGet_ReturnsEnvironmentValue verifies that Get returns the value
// from the environment when the variable is set.
func TestGet_ReturnsEnvironmentValue(t *testing.T) {
	key := "TEST_ENV_KEY"
	expected := "value-from-env"

	t.Setenv(key, expected)

	got := Get(key, "default-value")

	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

// TestGet_ReturnsDefaultValueWhenUnset verifies that Get returns the
// default value when the environment variable is not set.
func TestGet_ReturnsDefaultValueWhenUnset(t *testing.T) {
	key := "NON_EXISTENT_ENV_KEY"
	defaultValue := "default-value"

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset env: %v", err)
	}

	got := Get(key, defaultValue)

	if got != defaultValue {
		t.Fatalf("expected %s, got %s", defaultValue, got)
	}
}

// TestGet_ReturnsDefaultValueWhenEmpty verifies that Get returns the
// default value when the environment variable is set to an empty string.
//
// This ensures empty environment values are treated as "not configured".
func TestGet_ReturnsDefaultValueWhenEmpty(t *testing.T) {
	key := "EMPTY_ENV_KEY"
	defaultValue := "default-value"

	t.Setenv(key, "")

	got := Get(key, defaultValue)

	if got != defaultValue {
		t.Fatalf("expected %s, got %s", defaultValue, got)
	}
}

// TestGet_IsDeterministic verifies that repeated calls return the same result
// given the same environment state.
func TestGet_IsDeterministic(t *testing.T) {
	key := "DETERMINISTIC_KEY"
	value := "stable-value"

	t.Setenv(key, value)

	for i := 0; i < 5; i++ {
		got := Get(key, "default")

		if got != value {
			t.Fatalf("iteration %d: expected %s, got %s", i, value, got)
		}
	}
}

// TestFirstNonEmpty_ReturnsFirstValue verifies that FirstNonEmpty returns
// the first value when all inputs are non-empty.
func TestFirstNonEmpty_ReturnsFirstValue(t *testing.T) {
	got := FirstNonEmpty("a", "b", "c")

	if got != "a" {
		t.Fatalf("expected a, got %s", got)
	}
}

// TestFirstNonEmpty_SkipsEmptyValues verifies that FirstNonEmpty skips
// empty strings and returns the first non-empty value.
func TestFirstNonEmpty_SkipsEmptyValues(t *testing.T) {
	got := FirstNonEmpty("", "", "c", "d")

	if got != "c" {
		t.Fatalf("expected c, got %s", got)
	}
}

// TestFirstNonEmpty_AllEmpty verifies that FirstNonEmpty returns an empty
// string when all provided values are empty.
func TestFirstNonEmpty_AllEmpty(t *testing.T) {
	got := FirstNonEmpty("", "", "")

	if got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// TestFirstNonEmpty_SingleValue verifies that FirstNonEmpty returns the
// value when a single non-empty argument is provided.
func TestFirstNonEmpty_SingleValue(t *testing.T) {
	got := FirstNonEmpty("only")

	if got != "only" {
		t.Fatalf("expected only, got %s", got)
	}
}

// TestFirstNonEmpty_EmptyInput verifies that FirstNonEmpty returns an
// empty string when no arguments are provided.
func TestFirstNonEmpty_EmptyInput(t *testing.T) {
	got := FirstNonEmpty()

	if got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}
