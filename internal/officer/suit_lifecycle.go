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

package officer

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	goredis "github.com/redis/go-redis/v9"
)

// SuitLifecycleManager is responsible for emitting lifecycle events.
//
// It does not make lifecycle decisions.
// It only executes side-effects based on decisions made elsewhere.
type SuitLifecycleManager struct {
	Log logr.Logger
	RDB *goredis.Client
}

// emitSuitCloseRequested publishes a request to close a suit.
//
// This is an event emission only; no validation or decision logic is performed here.
func (m *SuitLifecycleManager) emitSuitCloseRequested(
	ctx context.Context,
	incidentID string,
	reason string,
) {
	payload := map[string]any{
		"incident_id": incidentID,
		"reason":      reason,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		m.Log.Error(err, "failed to encode suit close event payload")
		return
	}

	err = m.RDB.XAdd(ctx, &goredis.XAddArgs{
		Stream: SuitCloseRequestedStream,
		Values: map[string]any{
			"payload": string(data),
		},
	}).Err()

	if err != nil {
		m.Log.Error(err, "failed to publish suit close event")
	}

	m.Log.Info("suit close requested",
		"incident_id", incidentID,
		"reason", reason,
	)
}
