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

import "fmt"

// GroupAlreadyExistsError indicates that the Redis consumer group
// already exists and cannot be recreated.
type GroupAlreadyExistsError struct {
	Group  string
	Stream string
}

// Error implements the error interface.
func (e GroupAlreadyExistsError) Error() string {
	return fmt.Sprintf("redis: consumer group %q already exists on stream %q", e.Group, e.Stream)
}
