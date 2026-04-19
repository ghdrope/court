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
	"net/http"
	"time"
)

// Client handles communication with GitHub API.
type Client struct {
	httpClient *http.Client
	token      string
	repo       string // format: owner/repo
	baseURL    string
}

// NewClient creates a new GitHub API client.
func NewClient(token, repo string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		token:   token,
		repo:    repo,
		baseURL: "https://api.github.com",
	}
}
