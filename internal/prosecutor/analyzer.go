package prosecutor

// Analyzer defines how IncidentReports are evaluated.
//
// Implementation may apply simple rules, heuristics,
// or advanced AI/ML-based analysis (e.g., RAG pipelines).
type Analyzer interface {
	Analyze(r IncidentReport) string
}
