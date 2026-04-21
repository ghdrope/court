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

package env

import "fmt"

// MissingEnvError is returned when a required environment variable
// is not set or is empty.
//
// It is used to enforce strict configuration requirements for
// critical runtime dependencies.
type MissingEnvError struct {
	Key string
}

// Error implements the error interface.
func (e MissingEnvError) Error() string {
	return fmt.Sprintf("required environment variable %s is not set", e.Key)
}

// Require validates that an environment variable is present and non-empty.
//
// It returns a MissingEnvError when the value is empty.
func Require(key, value string) error {
	if value == "" {
		return MissingEnvError{Key: key}
	}
	return nil
}
