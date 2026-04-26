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

import (
	"testing"

	"github.com/ghdrope/court/internal/testhelper"
)

// TestBuildTitle ensures a valid incident is converted into a human-readable title.
func TestBuildTitle(t *testing.T) {
	t.Parallel()

	inc := testhelper.NewIncidentReport()

	got := buildTitle(inc)

	if got == "" {
		t.Fatalf("expected non-empty title")
	}

	if got != "🚨 api failed — cluster-1/ns-1" {
		t.Errorf("unexpected title: %s", got)
	}
}

// TestBuildTitle_Nil ensures nil input returns safe fallback title.
func TestBuildTitle_Nil(t *testing.T) {
	t.Parallel()

	got := buildTitle(nil)

	if got != "🚨 nil incident report" {
		t.Errorf("unexpected nil title: %s", got)
	}
}
