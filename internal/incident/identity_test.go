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

package incident

import (
	"testing"
)

// TestParseIncidentID ensures valid IDs are parsed correctly.
func TestParseIncidentID(t *testing.T) {
	t.Parallel()

	pod, ns, uid, err := ParseIncidentID("ns-1/api/uid-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pod != "ns-1" {
		t.Errorf("expected ns-1 as pod, got %s", pod)
	}

	if ns != "api" {
		t.Errorf("expected api as namespace, got %s", ns)
	}

	if uid != "uid-123" {
		t.Errorf("expected uid-123, got %s", uid)
	}
}

// TestParseIncidentID_Invalid ensures invalid format returns error.
func TestParseIncidentID_Invalid(t *testing.T) {
	t.Parallel()

	_, _, _, err := ParseIncidentID("invalid")

	if err == nil {
		t.Errorf("expected error for invalid id")
	}
}
