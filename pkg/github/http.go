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
	"io"
	"net/http"
)

// dojSON executes an HTTP request against the GitHub API using JSON encoding.
//
// It handles:
//   - request creation with context
//   - request creation with context
//   - JSON marshaling of the request body (if provided)
//   - standard GitHub headers
//   - authentication via bearer token (if configured)
//   - response status validation
//   - JSON decoding into the provided output struct (if not nil)
func (c *Client) doJSON(
	ctx context.Context,
	method string,
	url string,
	body any,
	out any,
) error {

	var reader io.Reader

	// Encode request body if present
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("github: marshal request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return fmt.Errorf("github: create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	// Add authentication if token is provided
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("github: request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Validate response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"github: API error: status=%d body=%s",
			resp.StatusCode,
			string(b),
		)
	}

	// Decode response if needed
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("github: decode response: %w", err)
		}
	}

	return nil
}
