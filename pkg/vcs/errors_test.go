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

import "testing"

// TestError_Error verifies the string formatting of VCS errors.
func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		message  string
		want     string
	}{
		{
			name:     "basic error formatting",
			provider: "github",
			message:  "rate limit exceeded",
			want:     "github: rate limit exceeded",
		},
		{
			name:     "empty provider",
			provider: "",
			message:  "something went wrong",
			want:     ": something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Error{
				Provider: tt.provider,
				Message:  tt.message,
			}

			if err.Error() != tt.want {
				t.Errorf("unexpected error string: got %q, want %q", err.Error(), tt.want)
			}
		})
	}
}
