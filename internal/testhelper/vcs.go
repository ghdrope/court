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

// FakeClient is a generic test double that simulates a VCS client.
//
// It is intentionally decoupled from any domain types in order to avoid
// import cycles. Instead, it uses generic `any` types for maximum flexibility
// in tests across different packages.
type FakeClient struct {
	// CreateIssueFunc allows overriding the behavior of CreateIssue.
	// If nil, a default successful response is returned.
	CreateIssueFunc func(ctx context.Context, owner, repo string, issue any) (any, error)

	// CloseIssueFunc allows overriding the behavior of CloseIssue.
	// If nil, a default successful response is returned.
	CloseIssueFunc func(ctx context.Context, issueURL string) (any, error)
}

// CreateIssue simulates the creation of an issue in a VCS provider.
//
// If CreateIssueFunc is defined, it is executed. Otherwise, a default
// successful response containing a placeholder URL is returned.
func (f *FakeClient) CreateIssue(
	ctx context.Context,
	owner, repo string,
	issue any,
) (any, error) {

	if f.CreateIssueFunc != nil {
		return f.CreateIssueFunc(ctx, owner, repo, issue)
	}

	return map[string]any{
		"url": "https://example.com/issues/1",
	}, nil
}

// CloseIssue simulates closing an issue in a VCS provider.
//
// If CloseIssueFunc is defined, it is executed. Otherwise, a default
// empty success response is returned.
func (f *FakeClient) CloseIssue(
	ctx context.Context,
	issueURL string,
) (any, error) {

	if f.CloseIssueFunc != nil {
		return f.CloseIssueFunc(ctx, issueURL)
	}

	return struct{}{}, nil
}
