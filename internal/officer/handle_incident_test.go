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
	"errors"
	"testing"

	"github.com/ghdrope/court/internal/incident"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// fakeRepo is a minimal test double for the IncidentRepository.
type fakeRepo struct {
	insertErr error
	called    bool
}

// Insert simulates the repository Insert method.
func (f *fakeRepo) Insert(ctx context.Context, r *incident.IncidentReport) error {
	f.called = true
	return f.insertErr
}

// TestHandleIncident_Success verifies the main execution flow.
//
// It ensures that:
//   - the repository Insert is called
//   - the function proceeds to Redis publishing
//   - an error is returned due to Redis being unreachable
func TestHandleIncident_Success(t *testing.T) {
	repo := &fakeRepo{}

	// Redis client pointing to an invalid address to force failure.
	rdb := goredis.NewClient(&goredis.Options{
		Addr: "localhost:0",
	})

	svc := &Service{
		Repo: repo,
		RDB:  rdb,
		Log:  zap.NewNop(),
	}

	report := &incident.IncidentReport{
		ID:        "incident-123",
		Cluster:   "c1",
		Namespace: "default",
		Pod:       "pod-1",
	}

	err := svc.HandleIncident(context.Background(), report)

	if err == nil {
		t.Fatal("expected error due to redis failure")
	}

	if !repo.called {
		t.Fatal("expected Insert to be called")
	}
}

// TestHandleIncident_NilReport verifies input validation.
//
// It ensures that passing a nil IncidentReport returns an error.
func TestHandleIncident_NilReport(t *testing.T) {
	svc := &Service{
		Log: zap.NewNop(),
	}

	err := svc.HandleIncident(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil incident")
	}
}

// TestHandleIncident_InsertFails verifies behavior when persistence fails.
//
// It ensures that:
//   - the error from the repository is returned
//   - Redis is not attempted (implicited via early return)
func TestHandleIncident_InsertFails(t *testing.T) {
	expectedErr := errors.New("db error")

	repo := &fakeRepo{
		insertErr: expectedErr,
	}

	svc := &Service{
		Repo: repo,
		Log:  zap.NewNop(),
	}

	report := &incident.IncidentReport{
		ID: "incident-123",
	}

	err := svc.HandleIncident(context.Background(), report)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

// TestHandleIncident_RedisFails verifies error propagation from Redis.
//
// It ensures that:
//   - Insert is called successfully
//   - Redis publishing fails
//   - the error is returned to the caller
func TestHandleIncident_RedisFails(t *testing.T) {
	repo := &fakeRepo{}

	// Invalid Redis address to simulate failure.
	rdb := goredis.NewClient(&goredis.Options{
		Addr: "localhost:0",
	})

	svc := &Service{
		Repo: repo,
		RDB:  rdb,
		Log:  zap.NewNop(),
	}

	report := &incident.IncidentReport{
		ID: "incident-123",
	}

	err := svc.HandleIncident(context.Background(), report)

	if err == nil {
		t.Fatal("expected redis error")
	}

	if !repo.called {
		t.Fatal("expected Insert to be called before Redis failure")
	}
}
