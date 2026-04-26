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

	"github.com/ghdrope/court/pkg/vcs"
)

// TestCreateIssue verifies successful issue creation.
func TestCreateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body["title"] != "Test" {
			t.Errorf("unexpected title")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"html_url": "https://github.com/x/y/issues/1",
		})
	}))
	defer server.Close()

	c := NewClient("token", WithBaseURL(server.URL))

	res, err := c.CreateIssue(context.Background(), "x", "y", vcs.Issue{
		Title: "Test",
		Body:  "Body",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.URL != "https://github.com/x/y/issues/1" {
		t.Errorf("unexpected URL: %q", res.URL)
	}
}

// TestCreateIssue_Error verifies API failure handling.
func TestCreateIssue_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient("token", WithBaseURL(server.URL))

	_, err := c.CreateIssue(context.Background(), "x", "y", vcs.Issue{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// TestCloseIssue verifies successful issue closure.
func TestCloseIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient("token", WithBaseURL(server.URL))

	_, err := c.CloseIssue(context.Background(),
		"https://github.com/owner/repo/issues/1",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCloseIssue_InvalidURL verifies validation of malformed URLs.
func TestCloseIssue_InvalidURL(t *testing.T) {
	c := NewClient("token")

	_, err := c.CloseIssue(context.Background(), "invalid-url")
	if err == nil {
		t.Fatal("expected error")
	}
}
