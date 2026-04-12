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
)

// newPatrolCommand starts the k8s controller.
func newPatrolCommand() *cobra.Command {

	var apiAddr string

	cmd := &cobra.Command{
		Use:  "patrol",
		Args: cobra.NoArgs,

		RunE: func(cmd *cobra.Command, _ []string) error {

			logger := ctrl.Log.WithName("officer")
			logger.Info("starting patrol controller")

			// connect to API server
			conn, err := grpc.NewClient(
				apiAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to API server: %w", err)
			}

			defer func() {
				if err = conn.Close(); err != nil {
					logger.Error(err, "failed to close gRPC connection")
				}
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
