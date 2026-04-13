package prosecutor

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Service coordinates fetching and analyzing IncidentReports.
// It represents the core processing unit of the Prosecutor component.
type Service struct {
	Fetcher  Fetcher
	Analyzer Analyzer
}

// NewService creates a new Prosecutor service with its dependencies.
func NewService(fetcher Fetcher, analyzer Analyzer) *Service {
	return &Service{
		Fetcher:  fetcher,
		Analyzer: analyzer,
	}
}

// ProcessBatch fetches IncidentReports and processes them through the Analyzer.
func (s *Service) ProcessBatch(ctx context.Context) error {
	reports, err := s.Fetcher.FetchIncidentReports(ctx)
	if err != nil {
		return fmt.Errorf("fetch reports: %w", err)
	}

	for _, r := range reports {
		comment := s.Analyzer.Analyze(r)

		zap.L().Info("incident analyzed",
			zap.String("id", r.ID),
			zap.String("result", comment))
	}

	return nil
}
