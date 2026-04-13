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
	"os"

	"github.com/ghdrope/court/internal/officer"
	grpcclient "github.com/ghdrope/court/internal/transport/grpc"
	incidentpb "github.com/ghdrope/court/proto/incident"

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

// defaultArchiveAddr provides zero-config.
const defaultArchiveAddr = "localhost:50052"

// newPatrolCommand starts the k8s controller loop.
func newPatrolCommand() *cobra.Command {

	var archiveAddr string

	cmd := &cobra.Command{
		Use:  "patrol",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {

			logger := ctrl.Log.WithName("officer")
			logger.Info("starting patrol controller")

			// Resolve Archive address
			if archiveAddr == "" {
				archiveAddr = os.Getenv("ARCHIVE_ADDR")
				if archiveAddr == "" {
					archiveAddr = defaultArchiveAddr
					logger.Info("using default archive address", "addr", archiveAddr)
				}
			}
			// Connect to Archive gRPC service
			conn, err := grpc.NewClient(
				archiveAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to archive: %w", err)
			}

			defer func() {
				if err = conn.Close(); err != nil {
					logger.Error(err, "failed to close gRPC connection")
				}
			}()

			// Create gRPC client
			pbClient := incidentpb.NewArchiveServiceClient(conn)

			// Wrap into domain client
			archiveClient := grpcclient.NewArchiveClient(pbClient)

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

			// Reconciler
			reconciler := &officer.PodReconciler{
				Client:  mgr.GetClient(),
				Log:     log.Log.WithName("reconciler"),
				Archive: archiveClient,
			}

			// Controller
			if err := ctrl.NewControllerManagedBy(mgr).
				For(&v1.Pod{}).
				Complete(reconciler); err != nil {
				return fmt.Errorf("failed to create controller: %w", err)
			}

			logger.Info("controller registered, starting manager")

			return mgr.Start(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&archiveAddr, "archive-addr", "", "default: env ARCHIVE_ADDR or localhost:50052")

	return cmd
}
