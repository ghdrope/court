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

package issue

import "testing"

// TestFormatLogs_Empty ensures empty logs return fallback message.
func TestFormatLogs_Empty(t *testing.T) {
	t.Parallel()

	got := formatLogs([]string{})

	if got != "<no logs available>" {
		t.Errorf("unexpected output for empty logs: %s", got)
	}
}

// TestFormatLogs_Content ensures logs are joined correctly.
func TestFormatLogs_Content(t *testing.T) {
	t.Parallel()

	logs := []string{
		"line1",
		"line2",
	}

	got := formatLogs(logs)

	if got != "line1\nline2" {
		t.Errorf("unexpected formatted logs: %s", got)
	}
}
