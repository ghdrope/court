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

package incident

import (
	"fmt"
	"strings"
)

// ParseIncidentID extracts namespace, pod and UID from an incident identifier.
//
// Expected format:
//
//	namespace/pod/uid
//
// UID is used only for uniqueness and may be ignored by consumers.
func ParseIncidentID(id string) (pod string, namespace string, uid string, err error) {
	parts := strings.Split(id, "/")

	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("invalid incident id format: %s", id)
	}

	pod = parts[0]
	namespace = parts[1]
	uid = parts[2]

	return pod, namespace, uid, nil
}
