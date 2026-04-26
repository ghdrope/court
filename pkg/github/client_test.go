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
	"testing"
	"time"
)

// TestNewClient_Defaults verifies that NewClient sets default values correctly.
func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("token")

	if c.token != "token" {
		t.Errorf("unexpected token: got %q, want %q", c.token, "token")
	}

	if c.baseURL != "https://api.github.com" {
		t.Errorf("unexpected baseURL: got %q", c.baseURL)
	}

	if c.httpClient == nil {
		t.Fatal("expected httpClient to be set")
	}

	if c.httpClient.Timeout != 10*time.Second {
		t.Errorf("unexpected timeout: got %v", c.httpClient.Timeout)
	}
}

// TestNewClient_Options verifies that options override defaults.
func TestNewClient_Options(t *testing.T) {
	customHTTP := &http.Client{Timeout: 1 * time.Second}

	c := NewClient(
		"token",
		WithHTTPClient(customHTTP),
		WithBaseURL("http://example.com"),
	)

	if c.httpClient != customHTTP {
		t.Error("expected custom http client")
	}

	if c.baseURL != "http://example.com" {
		t.Errorf("unexpected baseURL: got %q", c.baseURL)
	}
}
