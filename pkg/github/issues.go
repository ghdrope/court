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

package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ghdrope/court/pkg/vcs"
)

// ensure Client implements vcs.Client
var _ vcs.Client = (*Client)(nil)

// createIssueRequest represents the payload used to create a GitHub issue.
//
// Only the required fields are included. Additional fields (labels,
// assignees, etc.) can be added later without breaking the API.
type createIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// issueResponse represents the subset of fields returnevd by GitHub
// that are relevant for this client.
type issueResponse struct {
	HTMLURL string `json:"html_url"`
}

// CreateIssue creates a new issue in the specified GitHub repository.
//
// The repository is identified by owner and repo.
//
// It returns the public HTML URL of the created issue.
//
// This method implements the vcs.Client interface.
func (c *Client) CreateIssue(
	ctx context.Context,
	owner, repo string,
	issue vcs.Issue,
) (vcs.IssueResult, error) {

	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, owner, repo)

	payload := createIssueRequest{
		Title: issue.Title,
		Body:  issue.Body,
	}

	var resp issueResponse

	if err := c.doJSON(ctx, http.MethodPost, endpoint, payload, &resp); err != nil {
		return vcs.IssueResult{}, fmt.Errorf("github create issue failed: %w", err)
	}

	return vcs.IssueResult{
		URL: resp.HTMLURL,
	}, nil
}

// CloseIssue closes an issue.
func (c *Client) CloseIssue(ctx context.Context, issueURL string) (vcs.CloseResult, error) {
	// expected:
	// https://github.com/{owner}/{repo}/issues/{number}

	parts := strings.Split(strings.Trim(issueURL, "/"), "/")
	if len(parts) < 5 {
		return vcs.CloseResult{}, fmt.Errorf("invalid issue URL format")
	}

	owner := parts[len(parts)-4]
	repo := parts[len(parts)-3]
	number := parts[len(parts)-1]

	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues/%s",
		c.baseURL,
		owner,
		repo,
		number,
	)

	payload := map[string]any{
		"state": "closed",
	}

	if err := c.doJSON(ctx, http.MethodPatch, endpoint, payload, nil); err != nil {
		return vcs.CloseResult{}, err
	}

	return vcs.CloseResult{}, nil
}
