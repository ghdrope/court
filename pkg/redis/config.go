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

import "time"

// Config defines the configuration for a Redis Stream consumer group.
//
// It controls how messages are consumed from a Redis Stream.
type Config struct {
	// Stream is the Redis stream name.
	Stream string

	// Group is the consumer group name.
	Group string

	// Consumer is the consumer instance name.
	Consumer string

	// BatchSize defines how many messages are fetched per XREADGROUP call.
	BatchSize int

	// BlockTime defines how long the consumer waits for new messages.
	BlockTime time.Duration
}

// DefaultConfig returns a Config with safe default values.
//
// It should be used as a base configuration that can be customized.
func DefaultConfig(stream, group, consumer string) Config {
	return Config{
		Stream:    stream,
		Group:     group,
		Consumer:  consumer,
		BatchSize: 10,
		BlockTime: 5 * time.Second,
	}
}
