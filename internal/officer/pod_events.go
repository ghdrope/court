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

package officer

import (
	"context"

	"github.com/ghdrope/court/internal/incident"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fetchPodEvents retrieves Kubernetes Events associated with a Pod
// and converts them into the internal Incident event model.
//
// It is defensive by design because Events API is eventually consistent
// and may return partial or noisy data.
func (r *PodReconciler) fetchPodEvents(
	ctx context.Context,
	namespace string,
	pod *v1.Pod,
) ([]incident.K8sEvent, error) {

	if r.KubeClient == nil {
		return nil, nil
	}

	eventList, err := r.KubeClient.CoreV1().
		Events(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if eventList == nil {
		return nil, nil
	}

	result := make([]incident.K8sEvent, 0, len(eventList.Items))

	for _, e := range eventList.Items {

		// Only events for this Pod
		if e.InvolvedObject.Kind != "Pod" {
			continue
		}

		if e.InvolvedObject.UID != pod.UID {
			continue
		}

		if e.InvolvedObject.Name != pod.Name {
			continue
		}

		// Safety filter: ignore empty noise events
		if e.Type == "" && e.Reason == "" && e.Message == "" {
			continue
		}

		result = append(result, incident.K8sEvent{
			Type:    e.Type,
			Reason:  e.Reason,
			Message: e.Message,
		})
	}

	return result, nil
}
