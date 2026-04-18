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

package redis

import "testing"

// TestConfigInitialization verifies that the Config struct
// is correctly initialized with the provided values.
func TestConfigInitialization(t *testing.T) {
	cfg := Config{
		Stream:   "test-stream",
		Group:    "test-group",
		Consumer: "test-consumer",
	}

	if cfg.Stream != "test-stream" {
		t.Errorf("expected Stream to be %q, got %q", "test-stream", cfg.Stream)
	}

	if cfg.Group != "test-group" {
		t.Errorf("expected Group to be %q, got %q", "test-group", cfg.Group)
	}

	if cfg.Consumer != "test-consumer" {
		t.Errorf("expected Consumer to be %q, got %q", "test-consumer", cfg.Consumer)
	}
}

// TestConfigEmptyValues verifies that the Config struct
// can handle empty string values.
func TestConfigEmptyValues(t *testing.T) {
	cfg := Config{}

	if cfg.Stream != "" {
		t.Errorf("expected Stream to be empty, got %q", cfg.Stream)
	}

	if cfg.Group != "" {
		t.Errorf("expected Group to be empty, got %q", cfg.Group)
	}

	if cfg.Consumer != "" {
		t.Errorf("expected Consumer to be empty, got %q", cfg.Consumer)
	}
}
