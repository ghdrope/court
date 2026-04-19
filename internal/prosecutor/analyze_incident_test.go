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
	"errors"
	"testing"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/pkg/testhelper"
	"go.uber.org/zap"
)

// TestProcessIncident_NilIncident verifies nil input handling.
func TestProcessIncident_NilIncident(t *testing.T) {
	db, _ := testhelper.NewSQLMock(t)
	repo := incident.NewRepository(db)

	svc := &Service{
		Repo: repo,
		Log:  zap.NewNop(),
	}

	err := svc.ProcessIncident(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil incident")
	}
}

// TestProcessIncident_RepoError verifies repository failure.
func TestProcessIncident_RepoError(t *testing.T) {
	db, mock := testhelper.NewSQLMock(t)
	repo := incident.NewRepository(db)

	mock.ExpectExec("UPDATE incidents").
		WillReturnError(errors.New("db error"))

	svc := &Service{
		Repo: repo,
		Log:  zap.NewNop(),
	}

	inc := &incident.IncidentReport{
		ID: "test-id",
	}

	err := svc.ProcessIncident(context.Background(), inc)
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

// TestAnalyze_ReturnsExpectedValues verifies deterministic analyzer output.
func TestAnalyze_ReturnsExpectedValues(t *testing.T) {
	db, _ := testhelper.NewSQLMock(t)
	repo := incident.NewRepository(db)

	svc := &Service{
		Repo: repo,
		Log:  zap.NewNop(),
	}

	inc := &incident.IncidentReport{
		ID: "test-id",
	}

	commentary, repoURL := svc.analyze(inc)

	if commentary != "nothing to add" {
		t.Fatalf("unexpected commentary: %s", commentary)
	}

	if repoURL != "nothing to add" {
		t.Fatalf("unexpected repo URL: %s", repoURL)
	}
}
