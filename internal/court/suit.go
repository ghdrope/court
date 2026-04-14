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

package court

import (
	"context"
	"fmt"
	"time"
)

// CreateSuit inserts a new Suit referencing an IncidentReport.
//
// This is a formalization step: the incident already exists.
func (s *Service) CreateSuit(ctx context.Context, incidentID, cluster, namespace, pod string) error {

	query := `
		INSERT INTO suits (
			incident_id,
			cluster,
			namespace,
			pod,
			created_at
		)
		VALUES ($1,$2,$3,$4,$5)
	`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		incidentID,
		cluster,
		namespace,
		pod,
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("insert case: %w", err)
	}

	return nil
}
