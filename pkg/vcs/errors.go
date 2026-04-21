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

import "fmt"

// Error represents a generic error returned by a VCS provider.
type Error struct {
	// Provider is the name of the VCS provider.
	Provider string

	// Message describes the error returned by the provider.
	Message string
}

// Error returns the string representation of the error in the format "Provider: Message".
func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Provider, e.Message)
}
