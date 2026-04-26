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
	"net/url"
	"strings"
	"time"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/issue"
	"github.com/ghdrope/court/internal/suit"
	"go.uber.org/zap"
)

// CreateIssue creates a VCS issue and attaches it to the Suit.
func (s *Service) CreateIssue(
	ctx context.Context,
	inc *incident.IncidentReport,
	suitEntity *suit.Suit,
) error {

	if s.VCS == nil {
		s.Log.Warn("VCS client is nil - skipping issue creation")
		return nil
	}

	log := s.Log.With(
		zap.String("incident_id", inc.ID),
		zap.String("suit_id", suitEntity.ID),
	)

	owner, repo, err := parseRepo(inc.VCSRepoURL)
	if err != nil {
		log.Error("invalid repo url",
			zap.Error(err),
			zap.String("url", inc.VCSRepoURL),
		)
		return err
	}

	result, err := s.VCS.CreateIssue(
		ctx,
		owner,
		repo,
		issue.BuildFromIncidentReport(inc),
	)

	if err != nil {
		log.Error("failed to create VCS issue",
			zap.Error(err),
		)
		return err
	}

	suitEntity.VCSIssueURL = result.URL

	log.Info("VCS issue created",
		zap.String("issue_url", result.URL),
	)

	return nil
}

// parseRepo extracts owner/repo from a Git URL.
//
// Supports standard HTTPS GitHub URLs.
func parseRepo(raw string) (string, string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repo path: %s", u.Path)
	}

	return parts[0], parts[1], nil
}

// CloseIssue closes a VCS issue and updates Suit state.
//
// Future: extend to support multiple VCS providers.
func (s *Service) CloseIssue(
	ctx context.Context,
	suitEntity *suit.Suit,
) error {

	if s.VCS == nil {
		return nil
	}

	if suitEntity == nil {
		return fmt.Errorf("suit is nil")
	}

	if suitEntity.VCSIssueURL == "" {
		return fmt.Errorf("no VCS issue linked")
	}

	log := s.Log.With(
		zap.String("suit_id", suitEntity.ID),
		zap.String("issue_url", suitEntity.VCSIssueURL),
	)

	log.Info("closing VCS issue")

	_, err := s.VCS.CloseIssue(ctx, suitEntity.VCSIssueURL)
	if err != nil {
		log.Error("failed to close VCS issue", zap.Error(err))
		return err
	}

	now := time.Now()
	suitEntity.Status = suit.StatusClosed
	suitEntity.ClosedAt = &now

	log.Info("VCS issue closed")

	return nil
}
