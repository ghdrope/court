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
