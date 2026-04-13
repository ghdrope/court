package prosecutor

import "context"

// MockFetcher is a simple in-memory implementation of Fetcher.
// It returns static data for development and testing purposes.
type MockFetcher struct{}

// NewMockFetcher creates a new MockFetcher.
func NewMockFetcher() *MockFetcher {
	return &MockFetcher{}
}

// FetchIncidentReports returns a static list of IncidentReports.
func (m *MockFetcher) FetchIncidentReports(ctx context.Context) ([]IncidentReport, error) {
	return []IncidentReport{
		{ID: "incident-1"},
		{ID: "incident-2"},
	}, nil
}
