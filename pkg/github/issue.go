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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateIssueRequest represents GitHub issue payload.
type CreateIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// CreateIssue creates a new GitHub issue.
func (c *Client) CreateIssue(ctx context.Context, title, body string) error {
	url := fmt.Sprintf("%s/repos/%s/issues", c.baseURL, c.repo)

	payload := CreateIssueRequest{
		Title: title,
		Body:  body,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal issue: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("github request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	return nil
}
