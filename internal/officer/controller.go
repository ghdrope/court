package officer

import (
	"context"

	"github.com/ghdrope/court/internal/incident"
	"github.com/go-logr/logr"
	goredis "github.com/redis/go-redis/v9"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IncidentService interface {
	HandleIncident(ctx context.Context, r *incident.IncidentReport) error
}

type PodReconciler struct {
	client.Client

	KubeClient kubernetes.Interface
	Log        logr.Logger

	Service     IncidentService
	SuitManager *SuitLifecycleManager

	Cluster string

	ImageMetadataProvider ImageMetadataProvider
	RDB                   *goredis.Client

	RecoveryHints []RecoveryHint
}

func (r *PodReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {

	logger := r.Log.WithValues("ns", req.Namespace, "name", req.Name)

	if req.Namespace == "court" {
		return ctrl.Result{}, nil
	}

	r.processRecoveryHints(ctx)

	var pod v1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if shouldIgnorePod(&pod) {
		return ctrl.Result{}, nil
	}

	// 1. FAST DETECTION (no logs)
	containersMetadata := DetectContainerIssues(ctx, r.KubeClient, &pod)

	shouldHandle := len(containersMetadata) > 0 || isPodFailing(&pod, containersMetadata)

	if !shouldHandle {
		r.reconcileOpenSuits(ctx)
		return ctrl.Result{}, nil
	}

	// 2. ENSURE WE ALWAYS HAVE EVIDENCE BASE
	if len(containersMetadata) == 0 {
		containersMetadata = []incident.ContainerMetadata{
			{
				Container: firstContainer(&pod),
				ImageName: firstImage(&pod),
				Reason:    "non-k8s-signal incident",
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

func firstContainer(pod *v1.Pod) string {
	if len(pod.Spec.Containers) == 0 {
		return "unknown"
	}
	return pod.Spec.Containers[0].Name
}

func firstImage(pod *v1.Pod) string {
	if len(pod.Spec.Containers) == 0 {
		return "unknown"
	}
	return pod.Spec.Containers[0].Image
}

// ======================================================
// 1. HINT CONSUMPTION (single-use per cycle)
// ======================================================
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

// ======================================================
// 3. SUIT RECONCILIATION LOOP (always active truth engine)
// ======================================================
func (r *PodReconciler) reconcileOpenSuits(ctx context.Context) {

	logger := r.Log.WithName("suit-reconciliation")

	suits, err := r.Service.(*Service).SuitRepo.ListOpen(ctx)
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

		// HARD RULE: ignore system namespace
		if ns == "court" {
			continue
		}

		pod, err := r.KubeClient.CoreV1().
			Pods(ns).
			Get(ctx, podName, metav1.GetOptions{})

		// pod deleted → close suit
		if err != nil {
			r.SuitManager.emitSuitCloseRequested(
				ctx,
				suit.IncidentID,
				"pod_deleted",
			)
			continue
		}

		containersMetadata := DetectContainerIssues(ctx, r.KubeClient, pod)

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
