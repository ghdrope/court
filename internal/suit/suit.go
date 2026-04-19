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

package suit

import "time"

// Status represents the lifecycle state of a Suit.
type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

// Suit represents a wrapper around an Incident.
//
// It does NOT dulicate the IncidentReport. It only references it.
type Suit struct {
	ID         string
	IncidentID string
	Status     Status
	CreatedAt  time.Time
	ClosedAt   *time.Time
}
