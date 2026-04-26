package testhelper

import "github.com/ghdrope/court/internal/incident"

// NewIncidentReport creates a minimal valid IncidentReport for testing.
//
// It provides default values that can be overridden per test when needed.
func NewIncidentReport() *incident.IncidentReport {
	return &incident.IncidentReport{
		Cluster:    "cluster-1",
		Namespace:  "ns-1",
		Pod:        "api",
		VCSRepoURL: "https://github.com/example/repo",
	}
}

// NewIncidentWithEvents returns an IncidentReport with predefined events
// for testing event rendering logic.
func NewIncidentWithEvents() *incident.IncidentReport {
	inc := NewIncidentReport()

	inc.Events = []incident.K8sEvent{
		{Type: "info", Reason: "startup", Message: "started"},
		{Type: "warning", Reason: "cpu", Message: "high usage"},
	}

	return inc
}

// NewIncidentWithContainers returns an IncidentReport with container issues
// for testing container rendering logic.
func NewIncidentWithContainers() *incident.IncidentReport {
	inc := NewIncidentReport()

	inc.ContainersMetadata = []incident.ContainerMetadata{
		{
			Container: "api",
			ImageName: "nginx:latest",
			Logs:      []string{"error starting server"},
		},
	}

	return inc
}

// Contains reports whether substr is present in s.
func Contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || (StringIndex(s, substr) >= 0))
}

// StringIndex returns the index of the first occurrence of substr in s,
// or -1 if substr is not present.
func StringIndex(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
