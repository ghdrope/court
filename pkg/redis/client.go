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

import goredis "github.com/redis/go-redis/v9"

// New Client creates and returns a new Redis client.
func NewClient(opts *goredis.Options) *goredis.Client {
	return goredis.NewClient(opts)
}

// Config defines the configuration for a Redis Stream consumer group.
type Config struct {
	Stream   string // Stream name
	Group    string // Consumer group name
	Consumer string // Consumer instance name
}
