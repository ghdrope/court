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
	"github.com/ghdrope/court/pkg/vcs"
	"go.uber.org/zap"
)

// SuitRepository defines persistence operations for suits.
type SuitRepository interface {
	Insert(ctx context.Context, s *suit.Suit) error
	UpdateVCSInfo(ctx context.Context, s *suit.Suit) error
	GetByIncidentID(ctx context.Context, incidentID string) (*suit.Suit, error)
	Close(ctx context.Context, id string) error
}

// VCSClient defines a generic version control system interface.
//
// It is intentionally abstract to support multiple providers
// such as GitHub, GitLab, ...
type VCSClient interface {
	CreateIssue(ctx context.Context, owner, repo string, issue vcs.Issue) (vcs.IssueResult, error)
	CloseIssue(ctx context.Context, issueURL string) (vcs.CloseResult, error)
}

// Service orchestrates Suit and Issue lifecycle operations.
type Service struct {
	Repo SuitRepository
	VCS  VCSClient
	Log  *zap.Logger
}

// New creates a new Court service instance.
func New(repo SuitRepository, vcsClient VCSClient, log *zap.Logger) *Service {
	return &Service{
		Repo: repo,
		VCS:  vcsClient,
		Log:  log,
	}
}
