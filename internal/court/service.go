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

// SuitRepository defines persistence operations for Suit lifecycle.
//
// It abstracts database operations for creation, update, and closure
// of Suit entities tied to incidents.
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

// Service orchestrates Suit lifecycle and external VCS interactions.
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

// FakeSuitRepository is an in-memory implementation of SuitRepository used for testing.
type FakeSuitRepository struct {
	Suits map[string]*suit.Suit
}

// NewFakeSuitRepository creates an initialized in-memory repository for tests.
func NewFakeSuitRepository() *FakeSuitRepository {
	return &FakeSuitRepository{
		Suits: map[string]*suit.Suit{},
	}
}

// Insert stores a Suit in memory indexed by IncidentID.
func (f *FakeSuitRepository) Insert(ctx context.Context, s *suit.Suit) error {
	f.Suits[s.IncidentID] = s
	return nil
}

// UpdateVCSInfo updates an existing Suit in memory with new VCS metadata.
func (f *FakeSuitRepository) UpdateVCSInfo(ctx context.Context, s *suit.Suit) error {
	f.Suits[s.IncidentID] = s
	return nil
}

// GetByIncidentID retrieves a Suit by IncidentID.
// Returns nil if no Suit is found.
func (f *FakeSuitRepository) GetByIncidentID(ctx context.Context, incidentID string) (*suit.Suit, error) {
	if s, ok := f.Suits[incidentID]; ok {
		return s, nil
	}
	return nil, nil
}

// Close marks a Suit as closed in memory.
// If no matching Suit is found, the operation is a no-op.
func (f *FakeSuitRepository) Close(ctx context.Context, id string) error {
	for _, s := range f.Suits {
		if s.ID == id {
			s.Status = suit.StatusClosed
			return nil
		}
	}
	return nil
}

// FakeVCSClient is an in-memory stub implementation of VCSClient for tests.
//
// It records created and closed issues without performing external calls.
type FakeVCSClient struct {
	// CreatedIssues stores URLs of issues created during tests.
	CreatedIssues []string

	// ClosedIssues stores URLs of issues closed during tests.
	ClosedIssues []string
}

// CreateIssue simulates issue creation by returning a deterministic URL
// and recording it in memory.
func (f *FakeVCSClient) CreateIssue(ctx context.Context, owner, repo string, issue vcs.Issue) (vcs.IssueResult, error) {
	url := "https://vcs.test/" + owner + "/" + repo + "/issues/1"
	f.CreatedIssues = append(f.CreatedIssues, url)

	return vcs.IssueResult{
		URL: url,
	}, nil
}

// CloseIssue simulates closing an issue by recording the URL in memory.
func (f *FakeVCSClient) CloseIssue(ctx context.Context, issueURL string) (vcs.CloseResult, error) {
	f.ClosedIssues = append(f.ClosedIssues, issueURL)

	return vcs.CloseResult{}, nil
}
