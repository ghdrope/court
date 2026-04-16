package officer

import (
	"context"

	"github.com/ghdrope/court/internal/incident"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fetchPodEvents retrieves Kubernetes Events associated with a Pod
// and converts them into the internal Incident event model.
func (r *PodReconciler) fetchPodEvents(
	ctx context.Context,
	namespace string,
	podName string,
) ([]incident.K8sEvent, error) {

	eventList, err := r.KubeClient.CoreV1().
		Events(namespace).
		List(ctx, metav1.ListOptions{
			FieldSelector: "involvedObject.name=" + podName,
		})
	if err != nil {
		return nil, err
	}

	events := make([]incident.K8sEvent, 0, len(eventList.Items))

	for _, e := range eventList.Items {
		events = append(events, incident.K8sEvent{
			Type:    e.Type,
			Reason:  e.Reason,
			Message: e.Message,
		})
	}

	return events, nil
}
