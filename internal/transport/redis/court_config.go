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

package redisstream

// Stream names used by the Court service.
const (
	// IncidentCreatedStream is the stream where new incidents are published.
	IncidentCreatedStream = "incident.created"

	// SuitCloseRequestedStream is the stream where suit close requests are published.
	SuitCloseRequestedStream = "suit.close.requested"
)

// Consumer groups used by the Court service.
//
// Consumer groups define how Redis Stream messages are distributed
// among multiple consumers of the same service.
const (
	// CourtGroup is the shared consumer group used by the Court service.
	//
	// All Court consumers belong to this group to ensure coordinated
	// processing of stream messages.
	CourtGroup = "court-group"
)
