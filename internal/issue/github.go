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

// githubCopilotHintSection returns GitHub-specific guidance.
//
// This section is only included when the incident is linked to GitHub repositories.
func githubCopilotHintSection() string {
	return `
---

## 🤖 Copilot Hint

You can use GitHub Copilot Chat to analyze this issue and propose a fix:

- Ask: *"What is causing this failure?"*
- Ask: *"Propose a fix based on these logs"*
- Ask: *"Generate a patch or PR to prevent this crash"*
`
}
