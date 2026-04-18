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
	"testing"

	"github.com/ghdrope/court/internal/incident"
	"go.uber.org/zap"
)

// TestNewService verifies that the Service is correctly initialized
// with its dependencies properly assigned.
func TestNewService(t *testing.T) {
	repo := &incident.Repository{}
	logger := zap.NewNop()

	svc := New(repo, logger)

	if svc == nil {
		t.Fatal("expected service, got nil")
	}

	if svc.Repo != repo {
		t.Fatal("expected repo to be assigned correctly")
	}

	if svc.Log != logger {
		t.Fatal("expected logger to be assigned correctly")
	}
}

// TestServiceFieldsNotNil verifies that required dependencies
// are not nil after construction when provided.
func TestServiceFieldsNotNil(t *testing.T) {
	repo := &incident.Repository{}
	logger := zap.NewNop()

	svc := New(repo, logger)

	if svc.Repo == nil {
		t.Fatal("expected Repo to not be nil")
	}

	if svc.Log == nil {
		t.Fatal("expected Log to not be nil")
	}
}

// TestNewIsDeterministic verifies that multiple calls to New
// produce independent service instances.
func TestNewIsDeterministic(t *testing.T) {
	repo := &incident.Repository{}
	logger := zap.NewNop()

	svc1 := New(repo, logger)
	svc2 := New(repo, logger)

	if svc1 == svc2 {
		t.Fatal("expected different instances, got same pointer")
	}

	if svc1.Repo != svc2.Repo {
		t.Fatal("expected same repo reference")
	}

	if svc1.Log != svc2.Log {
		t.Fatal("expected same logger reference")
	}
}
