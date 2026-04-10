package controller

import (
	"context"

	"github.com/ghdrope/court/internal/inspector"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

// PodReconciler watches Pod objects and detects unhealthy transitions.
type PodReconciler struct {
	client.Client
	Log logr.Logger
}

// Reconcile runs whenever a Pod event occurs.
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := r.Log.WithValues(
		"namespace", req.Namespace,
		"name", req.Name,
	)

	var pod v1.Pod

	// Fetch latest Pod state
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	currentPhase := pod.Status.Phase

	// Detect transition: Running -> Failure
	if currentPhase == v1.PodFailed || currentPhase == v1.PodUnknown {
		logger.Info("pod in failure state", "phase", currentPhase)
	}

	// Detect container-level failures
	inspector.DetectContainerIssues(logger, &pod)

	return ctrl.Result{}, nil
}
