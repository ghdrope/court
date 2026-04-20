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
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

// IncidentService defines the contract required by the PodReconciler
// to persist and process incident reports.
//
// Implementations of this interface are responsible for handling
// the full lifecycle of an IncidentReport, including persistence,
// event emission, and downstream processing.
type IncidentService interface {
	HandleIncident(ctx context.Context, r *incident.IncidentReport) error
}

// PodReconciler reconciles Pod resources and produces IncidentReports
// for workloads that match known failure conditions.
type PodReconciler struct {
	client.Client

	KubeClient kubernetes.Interface
	Log        logr.Logger

	Service IncidentService

	Cluster string

	// ImageMetaProvider allows resolving OCI image labels
	// (e.g. org.opencontainers.image.source).
	//
	// This is required to enrich IncidentReports with repository metadata
	// when annotations are not present.
	ImageMetadataProvider ImageMetadataProvider
}

// Reconcile evaluates the current state of a Pod and determines whether
// it should generate an IncidentReport.
func (r *PodReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {

	logger := r.Log.WithValues(
		"namespace", req.Namespace,
		"name", req.Name,
	)

	var pod v1.Pod

	// Fetch latest Pod state
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Detect container-level issues
	containerIssues := DetectContainerIssues(ctx, r.KubeClient, &pod)

	// Fetch Kubernetes events for the Pod
	events, err := r.fetchPodEvents(ctx, pod.Namespace, &pod)
	if err != nil {
		logger.Error(err, "failed to fetch pod events")
		return ctrl.Result{}, err
	}

	// Determine if this Pod represents a problem worth reporting
	isProblem :=
		pod.Status.Phase == v1.PodFailed ||
			pod.Status.Phase == v1.PodUnknown ||
			len(containerIssues) > 0

	if !isProblem {
		return ctrl.Result{}, nil
	}

	// Resolve GitHub repository URL
	repoURL := resolveRepositoryURL(
		ctx,
		logger,
		&pod,
		r.ImageMetadataProvider,
	)

	if repoURL == "" {
		logger.Info("github repository could not be resolved for pod",
			"annotation", "court.dev/repository",
		)
	}
	// Build domain-level incident report
	report, err := incident.BuildFromPod(
		&pod,
		r.Cluster,
		repoURL,
		events,
		containerIssues,
	)
	if err != nil {
		logger.Error(err, "failed to build incident report")
		return ctrl.Result{}, err
	}

	logger.Info("incident detected",
		"id", report.ID,
	)

	// Delegate handling to service
	if err := r.Service.HandleIncident(ctx, &report); err != nil {
		logger.Error(err, "faile dto handle incident", "id", report.ID)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
