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

// TestDoJSON_Success verifies a successful JSON request/response cycle.
func TestDoJSON_Success(t *testing.T) {
	type response struct {
		Value string `json:"value"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("missing or invalid Authorization header")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{Value: "ok"})
	}))
	defer server.Close()

	c := NewClient("token", WithBaseURL(server.URL))

	var out response

	err := c.doJSON(context.Background(), http.MethodGet, server.URL, nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Value != "ok" {
		t.Errorf("unexpected response: got %q", out.Value)
	}
}

// TestDoJSON_APIError verifies handling of non-2xx responses.
func TestDoJSON_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	c := NewClient("")

	err := c.doJSON(context.Background(), http.MethodGet, server.URL, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

// TestDoJSON_InvalidJSON verifies decode failure.
func TestDoJSON_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid-json"))
	}))
	defer server.Close()

	c := NewClient("")

	var out map[string]any

	err := c.doJSON(context.Background(), http.MethodGet, server.URL, nil, &out)
	if err == nil {
		t.Fatal("expected decode error")
	}
}
