package incident

import (
	"fmt"
	"strings"
)

// ParseIncidentID extracts pod and namespace from an IncidentID.
//
// Expected format:
//
//	namespace/pod/uid
//
// Note:
// UID is ignored for recovery purposes.
func ParseIncidentID(id string) (pod string, namespace string, err error) {

	parts := strings.Split(id, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid incident id format: %s", id)
	}

	// namespace/pod/uid
	namespace = parts[0]
	pod = parts[1]

	return pod, namespace, nil
}
