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
	"testing"

	"github.com/ghdrope/court/internal/incident"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TestNew verifies that the Service constructor correctly assigns dependencies.
func TestNew(t *testing.T) {
	// Arrange
	repo := &incident.Repository{}
	rdb := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})
	logger := zap.NewNop()

	// Act
	svc := New(repo, rdb, logger)

	// Assert
	if svc == nil {
		t.Fatal("expected service to be initialized, got nil")
	}

	if svc.Repo != repo {
		t.Errorf("expected Repo to be %v, got %v", repo, svc.Repo)
	}

	if svc.RDB != rdb {
		t.Errorf("expected RDB to be %v, got %v", rdb, svc.RDB)
	}

	if svc.Log != logger {
		t.Errorf("expected Log to be %v, got %v", logger, svc.Log)
	}
}

// TestNew_WithNilDependencies verifies behavior when nil dependencies are provided.
func TestNew_WithNilDependencies(t *testing.T) {
	// Act
	svc := New(nil, nil, nil)

	// Assert
	if svc == nil {
		t.Fatal("expected service to be initialized even with nil dependencies")
	}

	if svc.Repo != nil {
		t.Errorf("expected Repo to be nil, got %v", svc.Repo)
	}

	if svc.RDB != nil {
		t.Errorf("expected RDB to be nil, got %v", svc.RDB)
	}

	if svc.Log != nil {
		t.Errorf("expected Log to be nil, got %v", svc.Log)
	}
}
