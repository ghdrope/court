package prosecutor

// StaticAnalyzer is a simple Analyzer implementation.
//
// it simulates analysis by always returning a static response.
type StaticAnalyzer struct{}

// NewStaticAnalyzer creates a new StaticAnalyzer.
func NewStaticAnalyzer() *StaticAnalyzer {
	return &StaticAnalyzer{}
}

// Analyze returns a static comment for any IncidentReport.
func (a *StaticAnalyzer) Analyze(r IncidentReport) string {
	return "nothing to add"
}
