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

	"github.com/ghdrope/court/internal/suit"
	"go.uber.org/zap"
)

type SuitRepository interface {
	Insert(ctx context.Context, s *suit.Suit) error
	GetByIncidentID(ctx context.Context, incidentID string) (*suit.Suit, error)
	Close(ctx context.Context, id string) error
}

// GitHubClient defines the behavior required to publish issues.
// This allows mocking and loose coupling with implementation.
type GitHubClient interface {
	CreateIssue(ctx context.Context, title, body string) (string, error)
}

// Service handles Suit lifecycle creation and updates.
type Service struct {
	Repo   SuitRepository
	GitHub GitHubClient
	Log    *zap.Logger
}

// New creates a new Court service instance.
func New(repo SuitRepository, gh GitHubClient, log *zap.Logger) *Service {
	return &Service{
		Repo:   repo,
		GitHub: gh,
		Log:    log,
	}
}
