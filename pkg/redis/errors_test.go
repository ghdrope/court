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

package redis

import "testing"

// TestGroupAlreadyExistsError verifies error formatting.
func TestGroupAlreadyExistsError(t *testing.T) {
	err := GroupAlreadyExistsError{
		Group:  "g",
		Stream: "s",
	}

	expected := `redis: consumer group "g" already exists on stream "s"`

	if err.Error() != expected {
		t.Errorf("unexpected error string: got %q, want %q", err.Error(), expected)
	}
}
