package controller

import (
	"context"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/inspector"
	"github.com/ghdrope/court/internal/router"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pb "github.com/ghdrope/court/proto/incident"
	ctrl "sigs.k8s.io/controller-runtime"
)

// PodReconciler reconciles Pod resources and detects failure conditions,
// producing IncidentReports for unhealthy pods.
type PodReconciler struct {
	client.Client
	Log logr.Logger

	API *router.APIClient
}

// Reconcile is triggered on Pod events and evaluates the current Pod state.
// When a failure condition or container issue is detected, it builds an
// incidentReport representing the observed problem.
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := r.Log.WithValues(
		"namespace", req.Namespace,
		"name", req.Name,
	)

	var pod v1.Pod

	// Fetch the latest state of the Pod from the API server.
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Inspect container-level issues such as CrashLoopBackOff or OOMKilled.
	containerIssues := inspector.DetectContainerIssues(&pod)

	isProblem :=
		pod.Status.Phase == v1.PodFailed ||
			pod.Status.Phase == v1.PodUnknown ||
			len(containerIssues) > 0

	if !isProblem {
		return ctrl.Result{}, nil
	}

	// Build a structured incident report from the Pod state.
	report := incident.BuildFromPod(
		&pod,
		containerIssues,
		[]string{}, // logs later (TBD)
	)

	pbReport := &pb.IncidentReport{
		EventId:   report.EventID,
		PodName:   report.PodName,
		Namespace: report.Namespace,
		Phase:     string(report.Phase),
		Reason:    report.Reason,
		Logs:      report.Logs,
	}

	logger.Info("sending incident", "event_id", report.EventID)

	if err := r.API.Send(ctx, pbReport); err != nil {
		logger.Error(err, "failed to send incident")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
