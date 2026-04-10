package main

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// PodReconciler watches Pod objects and detects unhealthy transitions.
type PodReconciler struct {
	client.Client
	Log logr.Logger
}

// newPatrolCommand starts the k8s controller.
func newPatrolCommand() *cobra.Command {

	return &cobra.Command{
		Use:  "patrol",
		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, _ []string) error {

			ctrl.SetLogger(ctrlzap.New())

			logger := ctrl.Log.WithName("officer")
			logger.Info("starting patrol controller")

			// Register Kubernetes API scheme
			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))

			// Build K8s config
			config := ctrl.GetConfigOrDie()

			// Create controller manager
			mgr, err := ctrl.NewManager(config, ctrl.Options{
				Scheme: scheme,
			})
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}

			// Create reconciler
			reconciler := &PodReconciler{
				Client: mgr.GetClient(),
				Log:    log.Log.WithName("reconciler"),
			}

			// Register Pod controller with manager
			if err := ctrl.NewControllerManagedBy(mgr).
				For(&v1.Pod{}).
				Complete(reconciler); err != nil {
				return fmt.Errorf("failed to create controller: %w", err)
			}

			logger.Info("controller registered, starting manager")

			return mgr.Start(cmd.Context())
		},
	}
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
	detectContainerIssues(logger, &pod)

	return ctrl.Result{}, nil
}

// detectContainerIssues inspects container statuses for real runtime failures
func detectContainerIssues(log logr.Logger, pod *v1.Pod) {

	for _, cs := range pod.Status.ContainerStatuses {

		// Waiting states
		if cs.State.Waiting != nil {
			switch cs.State.Waiting.Reason {
			case "CrashLoopBackOff", "ImagePullBackOff":
				log.Info("container issue detected",
					"container", cs.Name,
					"reason", cs.State.Waiting.Reason,
				)
			}
		}

		// Container terminated abnormally
		if cs.State.Terminated != nil {
			if cs.State.Terminated.Reason == "OOMKilled" {
				log.Info("container killed due to OOM",
					"container", cs.Name,
				)
			}
		}
	}
}
