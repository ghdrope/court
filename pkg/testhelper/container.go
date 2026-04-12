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

import v1 "k8s.io/api/core/v1"

// NewContainerStatusWaiting creates a container in Waiting state.
func NewContainerStatusWaiting(name, reason string) v1.ContainerStatus {
	return v1.ContainerStatus{
		Name: name,
		State: v1.ContainerState{
			Waiting: &v1.ContainerStateWaiting{
				Reason: reason,
			},
		},
	}
}

// NewContainerStatusTerminated creates a container in Terminated state.
func NewContainerStatusTerminated(name, reason string) v1.ContainerStatus {
	return v1.ContainerStatus{
		Name: name,
		State: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				Reason: reason,
			},
		},
	}
}
