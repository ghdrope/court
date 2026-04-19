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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCreateIssue_Success verifies that a GitHub issue is successfully created.
func TestCreateIssue_Success(t *testing.T) {

	var receivedRequest CreateIssueRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/repos/owner/repo/issues" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing or invalid Authorization header")
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content type")
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedRequest); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient("test-token", "owner/repo")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.CreateIssue(context.Background(), "test-title", "test-body")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if receivedRequest.Title != "test-title" {
		t.Fatalf("expected title 'test-title', got %s", receivedRequest.Title)
	}

	if receivedRequest.Body != "test-body" {
		t.Fatalf("expected body 'test-body', got %s", receivedRequest.Body)
	}
}

// TestCreateIssue_HTTPError verifies that HTTP errors are propagated.
func TestCreateIssue_HTTPError(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-token", "owner/repo")
	client.httpClient = server.Client()
	client.baseURL = server.URL

	err := client.CreateIssue(context.Background(), "title", "body")

	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

// TestCreateIssue_RequestError verifies request creation failure.
func TestCreateIssue_RequestError(t *testing.T) {

	client := NewClient("test-token", "owner/repo")

	client.baseURL = "://invalid-url"

	err := client.CreateIssue(context.Background(), "title", "body")

	if err == nil {
		t.Fatal("expected error for invalid request")
	}
}
