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
	"github.com/go-logr/logr"
	goredis "github.com/redis/go-redis/v9"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodReconciler reconciles Pod state into incidents and suit lifecycle events.
type PodReconciler struct {
	client.Client

	KubeClient kubernetes.Interface
	Log        logr.Logger

	Service  IncidentService
	SuitRepo SuitRepository

	SuitManager *SuitLifecycleManager

	Cluster string

	ImageMetadataProvider ImageMetadataProvider
	RDB                   *goredis.Client

	RecoveryHints []RecoveryHint
}

// Reconcile processes Pod events and maintains incident/suit state.
//
// It performs:
//   - recovery hint consumption
//   - incident detection and reporting
//   - suit lifecycle reconciliation
func (r *PodReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {

	logger := r.Log.WithValues("ns", req.Namespace, "name", req.Name)

	// Ignore system namespace
	if req.Namespace == "court" {
		return ctrl.Result{}, nil
	}

	// Consume recovery hints (one-shot)
	r.processRecoveryHints(ctx)

	var pod v1.Pod
	err := r.Get(ctx, req.NamespacedName, &pod)

	// Pod no longer exists -> treat as deletion event
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("pod not found (treated as deletion)")

			// Trigger lifecycle reconciliation to close suits if needed
			r.reconcileOpenSuits(ctx)

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Skip early startup noise
	if shouldIgnorePod(&pod) {
		return ctrl.Result{}, nil
	}

	// Fast detection (no log inspection)
	containersMetadata := DetectContainersMetadata(ctx, r.KubeClient, &pod)

	shouldHandle := len(containersMetadata) > 0 ||
		isPodFailing(&pod, containersMetadata)

	// No incident -> still reconcile suits
	if !shouldHandle {
		r.reconcileOpenSuits(ctx)
		return ctrl.Result{}, nil
	}

	// Ensure minimal evidence base
	if len(containersMetadata) == 0 {
		containersMetadata = []incident.ContainerMetadata{
			{
				Container: firstContainer(&pod),
				ImageName: firstImage(&pod),
				Reason:    "non_k8s_signal",
			},
		}
	}

	repoURL := resolveRepositoryURL(ctx, logger, &pod, r.ImageMetadataProvider)
	if repoURL == "" {
		return ctrl.Result{}, nil
	}

	events, err := r.fetchPodEvents(ctx, pod.Namespace, &pod)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Defensive fallback (log enrichment path)
	if len(containersMetadata) == 0 {
		containersMetadata = []incident.ContainerMetadata{
			{
				Container: firstContainer(&pod),
				ImageName: firstImage(&pod),
				Reason:    "log_enrichment_fallback",
			},
		}
	}

	containersMetadata = EnrichContainersMetadataWithLogs(
		ctx,
		r.KubeClient,
		&pod,
		containersMetadata,
	)

	report, err := incident.BuildFromPod(
		&pod,
		r.Cluster,
		repoURL,
		events,
		containersMetadata,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.Service.HandleIncident(ctx, &report)
}

// firstContainer returns the name of the first container in the Pod spec.
//
// If the Pod has no containers defined, it returns "unknown" as a safe fallback.
func firstContainer(pod *v1.Pod) string {
	if len(pod.Spec.Containers) == 0 {
		return "unknown"
	}
	return pod.Spec.Containers[0].Name
}

// firstImage returns the image reference of the first container in the Pod spec.
//
// If the Pod has no containers defined, it returns "unknown" as a safe fallback.
func firstImage(pod *v1.Pod) string {
	if len(pod.Spec.Containers) == 0 {
		return "unknown"
	}
	return pod.Spec.Containers[0].Image
}

// processRecoveryHints consumes recovery hints once per reconciliation cycle.
func (r *PodReconciler) processRecoveryHints(ctx context.Context) {

	logger := r.Log.WithName("recovery-hints")

	if len(r.RecoveryHints) == 0 {
		return
	}

	hints := r.RecoveryHints
	r.RecoveryHints = nil

	for _, h := range hints {
		logger.Info("consuming recovery hint",
			"incident_id", h.IncidentID,
			"current_uid", h.CurrentUID,
		)
	}
}

// reconcileOpenSuits evaluates all open suits against current cluster state.
//
// It acts as a continuous truth reconciliation loop ensuring that
// suits are closed when their conditions are no longer valid.
func (r *PodReconciler) reconcileOpenSuits(ctx context.Context) {

	if r.SuitRepo == nil || r.SuitManager == nil {
		return
	}

	logger := r.Log.WithName("suit-reconciliation")

	suits, err := r.SuitRepo.ListOpen(ctx)
	if err != nil {
		logger.Error(err, "failed to list open suits")
		return
	}

	for _, suit := range suits {

		podName, ns, expectedUID, err := incident.ParseIncidentID(suit.IncidentID)
		if err != nil {
			logger.Error(err, "invalid incident id",
				"incident_id", suit.IncidentID,
			)
			continue
		}

		// Ignore system namespace
		if ns == "court" {
			continue
		}

		pod, err := r.KubeClient.CoreV1().
			Pods(ns).
			Get(ctx, podName, metav1.GetOptions{})

		// Pod deleted -> close suit
		if err != nil {
			r.SuitManager.emitSuitCloseRequested(
				ctx,
				suit.IncidentID,
				"pod_deleted",
			)
			continue
		}

		containersMetadata := DetectContainersMetadata(ctx, r.KubeClient, pod)

		shouldClose, reason := EvaluateSuitClosure(
			pod,
			expectedUID,
			containersMetadata,
		)

		if shouldClose {

			r.SuitManager.emitSuitCloseRequested(
				ctx,
				suit.IncidentID,
				reason,
			)

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
}
