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
