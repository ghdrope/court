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

import (
	"testing"
	"time"
)

// TestDefaultConfig verifies that DefaultConfig returns expected defaults.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("stream", "group", "consumer")

	if cfg.Stream != "stream" {
		t.Errorf("unexpected Stream: got %q", cfg.Stream)
	}

	if cfg.Group != "group" {
		t.Errorf("unexpected Group: got %q", cfg.Group)
	}

	if cfg.Consumer != "consumer" {
		t.Errorf("unexpected Consumer: got %q", cfg.Consumer)
	}

	if cfg.BatchSize != 10 {
		t.Errorf("unexpected BatchSize: got %d, want 10", cfg.BatchSize)
	}

	if cfg.BlockTime != 5*time.Second {
		t.Errorf("unexpected BlockTime: got %v", cfg.BlockTime)
	}
}
