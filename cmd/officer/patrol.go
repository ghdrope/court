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
	redisstream "github.com/ghdrope/court/internal/transport/redis"
	"github.com/redis/go-redis/v9"

	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// defaultArchiveAddr provides zero-config.
const defaultRedisAddr = "localhost:6379"

// newPatrolCommand starts the k8s controller loop.
func newPatrolCommand() *cobra.Command {

	var redisAddr string
	var clusterName string

	cmd := &cobra.Command{
		Use:  "patrol",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {

			logger := ctrl.Log.WithName("officer")
			logger.Info("starting patrol controller")

			// Resolve Cluster name
			if clusterName == "" {
				clusterName = os.Getenv("CLUSTER_NAME")
				if clusterName == "" {
					return fmt.Errorf("cluster name must be set via --cluster or CLUSTER_NAME")
				}
			}

			// Resolve Redis address
			if redisAddr == "" {
				redisAddr = os.Getenv("REDIS_ADDR")
				if redisAddr == "" {
					redisAddr = defaultRedisAddr
					logger.Info("using default redis address", "redis-addr", redisAddr)
				}
			}

			// Create Redis base client
			rdb := redis.NewClient(&redis.Options{
				Addr: redisAddr,
			})
			baseClient := redisstream.NewClient(rdb)

			// Domain-specific stream client
			incidentClient := redisstream.IncidentStreamClient{
				Client: baseClient,
			}

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
				Archive: &incidentClient,
				Cluster: clusterName,
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

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")
	cmd.Flags().StringVar(&clusterName, "cluster", "", "Cluster name")

	return cmd
}
