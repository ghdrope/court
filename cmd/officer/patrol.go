package main

import (
	"fmt"

	"github.com/ghdrope/court/internal/controller"
	"github.com/ghdrope/court/internal/router"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// newPatrolCommand starts the k8s controller.
func newPatrolCommand() *cobra.Command {

	var apiAddr string

	cmd := &cobra.Command{
		Use:  "patrol",
		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, _ []string) error {

			ctrl.SetLogger(ctrlzap.New())

			logger := ctrl.Log.WithName("officer")
			logger.Info("starting patrol controller")

			// connect to API server
			conn, err := grpc.NewClient(
				apiAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return err
			}
			defer func() {
				_ = conn.Close()
			}()

			apiClient := router.NewAPIClient(conn)

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
			reconciler := &controller.PodReconciler{
				Client: mgr.GetClient(),
				Log:    log.Log.WithName("reconciler"),
				API:    apiClient,
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

	cmd.Flags().StringVar(&apiAddr, "api-addr", "localhost:50051", "API server gRPC address")

	return cmd
}
