package prosecutor

import "context"

// Fetcher defines how IncidentReports are retrieved from a source.
//
// Implementations may fetch from databases, message queues,
// APIs, or other event streams.
type Fetcher interface {
	FetchIncidentReports(ctx context.Context) ([]IncidentReport, error)
}
