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
	"fmt"

	"github.com/ghdrope/court/internal/incident"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RecoveryHint signals that a Pod was recreated (UID changed).
//
// This allows the controller to maintain continuity across Pod recreation
// without prematurely closing the associated suit.
type RecoveryHint struct {
	IncidentID string
	Namespace  string
	PodName    string

	PreviousUID string
	CurrentUID  string
}

// RecoverOpenIncidents reconciles open suits against the current cluster state.
//
// It performs a full scan of open suits and applies the same closure rules
// used during controller reconciliation.
//
// Returns:
//   - RecoveryHints for UID transitions (pod recreation)
//   - error if persistence access fails
func (s *Service) RecoverOpenIncidents(
	ctx context.Context,
	cluster string,
	kube kubernetes.Interface,
) ([]RecoveryHint, error) {

	logger := s.Log.WithValues("phase", "recovery", "cluster", cluster)

	logger.Info("starting suit recovery")

	suits, err := s.SuitRepo.ListOpen(ctx)
	if err != nil {
		logger.Error(err, "failed to list open suits")
		return nil, fmt.Errorf("list open suits: %w", err)
	}

	logger.Info("open suits loaded", "count", len(suits))

	manager := &SuitLifecycleManager{
		Log: s.Log.WithName("lifecycle"),
		RDB: s.RDB,
	}

	var hints []RecoveryHint

	for _, suit := range suits {

		podName, ns, expectedUID, err := incident.ParseIncidentID(suit.IncidentID)
		if err != nil {
			logger.Error(err, "invalid incident id, closing defensively",
				"incident_id", suit.IncidentID,
			)

			manager.emitSuitCloseRequested(ctx, suit.IncidentID, "invalid_incident_id")
			continue
		}

		pod, err := kube.CoreV1().
			Pods(ns).
			Get(ctx, podName, metav1.GetOptions{})

		// POD DOES NOT EXIST → hard close
		if err != nil {
			manager.emitSuitCloseRequested(ctx, suit.IncidentID, "pod_deleted")
			continue
		}

		containersMetadata := DetectContainersMetadata(ctx, kube, pod)

		shouldClose, reason := EvaluateSuitClosure(
			pod,
			expectedUID,
			containersMetadata,
		)

		if shouldClose {

			// UID transition → emit hint instead of closing
			if reason == "pod_recreated" && expectedUID != "" {

				hints = append(hints, RecoveryHint{
					IncidentID:  suit.IncidentID,
					Namespace:   ns,
					PodName:     podName,
					PreviousUID: expectedUID,
					CurrentUID:  string(pod.UID),
				})

				logger.Info("pod recreation detected (hint emitted, suit kept open)",
					"incident_id", suit.IncidentID,
				)

				continue
			}

			// All other closure cases
			manager.emitSuitCloseRequested(ctx, suit.IncidentID, reason)

			logger.Info("suit closed",
				"incident_id", suit.IncidentID,
				"reason", reason,
			)

			continue
		}

		logger.V(1).Info("suit remains open",
			"incident_id", suit.IncidentID,
		)
	}

	logger.Info("recovery completed")

	return hints, nil
}
