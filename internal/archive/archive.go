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

package archive

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Archive provides access to persistence operations.
type Archive struct {
	DB        *sql.DB
	Publisher EventPublisher
}

// New creates a new Archive instance backed by PostgreSQL.
func New(db *sql.DB, publisher EventPublisher) *Archive {
	return &Archive{
		DB:        db,
		Publisher: publisher}
}

// WaitForDB retries database connectivity until success or context cancellation.
func WaitForDB(ctx context.Context, db *sql.DB) error {

	const maxRetries = 10
	const delay = 2 * time.Second

	var err error

	for i := 0; i < maxRetries; i++ {
		err = db.PingContext(ctx)
		if err == nil {
			zap.L().Info("database connection established")
			return nil
		}

		zap.L().Warn("waiting for database",
			zap.Int("attempt", i+1),
			zap.Error(err),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("database unavailable after %d attempts: %w", maxRetries, err)
}

// InitSchema ensures all required database tables exist.
//
// It is safe to call multiple times.
func (a *Archive) InitSchema(ctx context.Context) error {
	return initIncidentSchema(ctx, a.DB)
}
