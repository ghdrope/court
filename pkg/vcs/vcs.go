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

package vcs

import "context"

// Issue represents a generic issue across VCS providers.
type Issue struct {
	Title string
	Body  string
}

// IssueResult represents the result of an issue creation.
type IssueResult struct {
	URL string
}

// CloseResult represents the result of closing an issue.
type CloseResult struct{}

// Client defines the behavior supported by a VCS provider.
//
// Implementations may include GitHub, GitLab, etc.
type Client interface {
	// CreateIssue creates a new issue in the given repository.
	//
	// The repository is identified by owner and repo.
	CreateIssue(ctx context.Context, owner, repo string, issue Issue) (IssueResult, error)

	// CloseIssue closes an existing issue.
	CloseIssue(ctx context.Context, issueURL string) (CloseResult, error)
}
