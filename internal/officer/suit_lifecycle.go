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

// SuitLifecycleManager handles lifecycle side-effects only.
//
// It does NOT decide lifecycle state.
// It only emits lifecycle events.
type SuitLifecycleManager struct {
	Log logr.Logger
	RDB *goredis.Client
}

// EmitSuitCloseRequested publishes a suit closure request event.
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
		m.Log.Error(err, "failed to encode suit close event")
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
