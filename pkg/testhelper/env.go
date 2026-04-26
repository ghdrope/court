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

package testhelper

import "os"

// SetEnv sets an environment variable for testing purposes.
//
// It returns a cleanup function that restores the previous value.
func SetEnv(key, value string) func() {
	old, existed := os.LookupEnv(key)

	if err := os.Setenv(key, value); err != nil {
		panic(err)
	}

	return func() {
		if !existed {
			if err := os.Unsetenv(key); err != nil {
				panic(err)
			}
			return
		}

		if err := os.Setenv(key, old); err != nil {
			panic(err)
		}
	}
}

// UnsetEnv removes an environment variable for testing purposes.
//
// It returns a cleanup function that restores the previous value if it existed.
func UnsetEnv(key string) func() {
	old, existed := os.LookupEnv(key)

	if err := os.Unsetenv(key); err != nil {
		panic(err)
	}

	return func() {
		if existed {
			if err := os.Setenv(key, old); err != nil {
				panic(err)
			}
		}
	}
}
