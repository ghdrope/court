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

package vcs

import (
	"context"
	"testing"

	testhelper "github.com/ghdrope/court/internal/testhelper"
)

// TestFakeClient_CreateIssue verifies that CreateIssue returns a valid issue map.
func TestFakeClient_CreateIssue(t *testing.T) {
	client := &testhelper.FakeClient{}

	res, err := client.CreateIssue(
		context.Background(),
		"owner",
		"repo",
		map[string]any{
			"title": "Test issue",
			"body":  "This is a test",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fake returns map[string]any by default
	result, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("unexpected type for result: %T", res)
	}

	expected := "https://example.com/issues/1"

	if result["url"] != expected {
		t.Errorf("unexpected URL: got %v, want %q", result["url"], expected)
	}
}

// TestFakeClient_CloseIssue verifies that CloseIssue executes successfully.
func TestFakeClient_CloseIssue(t *testing.T) {
	client := &testhelper.FakeClient{}

	_, err := client.CloseIssue(
		context.Background(),
		"https://example.com/issues/1",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
