package testhelper

import (
	v1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
)

// NewTestPod creates a basic Pod object for testing.
func NewTestPod(namespace, name string, phase v1.PodPhase) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}
}

// NewPodWithStatuses creates a pod with given container statuses.
func NewPodWithStatuses(statuses []v1.ContainerStatus) *v1.Pod {
	return &v1.Pod{
		Status: v1.PodStatus{
			ContainerStatuses: statuses,
		},
	}
}
