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

package prosecutor

import (
	"context"
	"fmt"

	"github.com/ghdrope/court/internal/incident"
)

// ProcessIncident performs post-processing over a stored IncidentReport.
//
// It performs analysis and updates the corresponding record in the
// Archive by setting the prosecutor_commentary field.
//
// Its operation must remain idempotent at the database level, performing
// an UPDATE based on the incident's unique event_id.
func (s *Service) ProcessIncident(
	ctx context.Context,
	r *incident.IncidentReport,
) error {

	if r == nil {
		return fmt.Errorf("incident is nil")
	}

	// TBD
	commentary := "nothing to add"

	query := `
		UPDATE incidents
		SET prosecutor_commentary = $1
		WHERE event_id = $2
	`

	res, err := s.DB.ExecContext(
		ctx,
		query,
		commentary,
		r.ID,
	)
	if err != nil {
		return fmt.Errorf("update incident commentary: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("incident not found for id=%s", r.ID)
	}

	return nil
}
