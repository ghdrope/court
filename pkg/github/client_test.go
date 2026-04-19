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
	"testing"
	"time"
)

// TestNewClient verifies that the GitHub client is correctly initialized.
func TestNewClient(t *testing.T) {

	token := "test-token"
	repo := "owner/repo"

	client := NewClient(token, repo)

	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.token != token {
		t.Fatalf("expected token %s, got %s", token, client.token)
	}

	if client.repo != repo {
		t.Fatalf("expected repo %s, got %s", repo, client.repo)
	}

	if client.httpClient == nil {
		t.Fatal("expected http client to be initialized")
	}

	expectedTimeout := 10 * time.Second
	if client.httpClient.Timeout != expectedTimeout {
		t.Fatalf("expected timeout %v, got %v", expectedTimeout, client.httpClient.Timeout)
	}
}

// TestNewClient_Deterministic verifies that multiple calls to NewClient
// return independent instances with identical configuration.
func TestNewClient_Deterministic(t *testing.T) {

	token := "test-token"
	repo := "owner/repo"

	c1 := NewClient(token, repo)
	c2 := NewClient(token, repo)

	if c1 == c2 {
		t.Fatal("expected different instances, got same pointer")
	}

	if c1.token != c2.token {
		t.Fatal("expected same token value")
	}

	if c1.repo != c2.repo {
		t.Fatal("expected same repo value")
	}

	if c1.httpClient == c2.httpClient {
		t.Fatal("expected different http client instances")
	}

	if c1.httpClient.Timeout != c2.httpClient.Timeout {
		t.Fatal("expected same timeout configuration")
	}
}
