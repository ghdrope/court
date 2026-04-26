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

// TestVCSSectionFor_GitHub ensures GitHub repositories
// return Copilot hint section.
func TestVCSSectionFor_GitHub(t *testing.T) {
	t.Parallel()

	inc := testhelper.NewIncidentReport()
	inc.VCSRepoURL = "https://github.com/example/repo"

	got := vcsSectionFor(inc)

	if got == "" {
		t.Fatalf("expected GitHub section, got empty string")
	}

	if got != githubCopilotHintSection() {
		t.Errorf("unexpected GitHub section output")
	}
}

// TestVCSSectionFor_Unknown ensures unknown providers return empty section.
func TestVCSSectionFor_Unknown(t *testing.T) {
	t.Parallel()

	inc := testhelper.NewIncidentReport()
	inc.VCSRepoURL = "https://gitlab.com/example/repo"

	got := vcsSectionFor(inc)

	if got != "" {
		t.Errorf("expected empty section for unknown provider, got: %s", got)
	}
}

// TestVCSSectionFor_Nil ensures nil incident returns empty section safely.
func TestVCSSectionFor_Nil(t *testing.T) {
	t.Parallel()

	got := vcsSectionFor(nil)

	if got != "" {
		t.Errorf("expected empty string for nil incident")
	}
}
