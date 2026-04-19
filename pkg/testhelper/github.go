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

package testhelper

import "context"

// GitHubMock is a reusable mock implementation of a GitHub client.
//
// It records invocation details and allows error injection for testing.
type GitHubMock struct {
	Called bool

	Title string
	Body  string

	URL string
	Err error
}

// CreateIssue records the call and returns a predefined error if set.
func (m *GitHubMock) CreateIssue(ctx context.Context, title, body string) (string, error) {
	m.Called = true
	m.Title = title
	m.Body = body

	if m.Err != nil {
		return "", m.Err
	}

	// deterministic fake URL for tests
	m.URL = "https://github.com/test/repo/issues/1"

	return m.URL, nil
}
