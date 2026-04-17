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

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/officer"
	"github.com/ghdrope/court/pkg/env"
	"github.com/ghdrope/court/pkg/postgres"
	"github.com/redis/go-redis/v9"

	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultClusterName = "local-cluster"
const defaultDsnAddr = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"
const defaultRedisAddr = "localhost:6379"

// newOfficerCommand initializes the Officer service.
//
// The Officer runs as a K8s controller responsible for:
//   - Watching Pod lifecycle events
//   - Detecting runtime failures
//   - Building IncidentReports from cluster state
//   - Persisting incidents into PostgreSQL
//   - Publishing incident events to Redis for downstream consumers
func newOfficerCommand() *cobra.Command {

	var redisAddr string
	var clusterName string

	cmd := &cobra.Command{
		Use:  "patrol",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {

			logger := ctrl.Log.WithName("officer")
			logger.Info("starting patrol controller")

			// Resolve cluster identity.
			// Used to uniquely identify the source Kubernetes environment.
			clusterName := env.Get("CLUSTER_NAME", defaultClusterName)

			// Resolve PostgreSQL connection string.
			// Used for persisting IncidentReports.
			dsn := env.Get("DATABASE_URL", defaultDsnAddr)

			// Resolve Redis address.
			// Used for emitting event notifications to downstream services.
			redisAddr := env.Get("REDIS_ADDR", defaultRedisAddr)

			// Initialize PostgreSQL connection.
			// This is required for persisting incident storage.
			db, err := postgres.Open(dsn)
			if err != nil {
				return fmt.Errorf("db open: %w", err)
			}

			// Ensure database readiness before controller startup.
			if err := postgres.PingWithRetry(cmd.Context(), db); err != nil {
				return fmt.Errorf("db not ready: %w", err)
			}

			// Initialize incident repository layer.
			// This is the only abstraction allowed to access the incidents table.
			repo := incident.NewRepository(db)

			// Ensure database schema exists before processing workloads.
			if err := repo.InitSchema(cmd.Context()); err != nil {
				return fmt.Errorf("init schema: %w", err)
			}

			// Initialize Redis client for event publishing.
			rdb := redis.NewClient(&redis.Options{
				Addr: redisAddr,
			})

			// Register Kubernetes API scheme for controller-runtime.
			scheme := runtime.NewScheme()
			utilruntime.Must(clientgoscheme.AddToScheme(scheme))

			// Create K8s controller manager.
			// Manages reconciliation loops and lifecycle of controllers.
			config := ctrl.GetConfigOrDie()

			mgr, err := ctrl.NewManager(config, ctrl.Options{
				Scheme: scheme,
			})
			if err != nil {
				return fmt.Errorf("failed to create manager: %w", err)
			}

			// Create Kubernetes client for direct API calls (logs, events, etc.)
			kubernetesClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}

			// Initialize Pod reconciler.
			// It detects failures and generates IncidentReports.
			reconciler := &officer.PodReconciler{
				Client:     mgr.GetClient(),
				KubeClient: kubernetesClient,
				Log:        log.Log.WithName("reconciler"),
				Cluster:    clusterName,
				Repo:       repo,
				RDB:        rdb,
			}

			// Register controller with manager.
			// Watches Pod resources and triggers reconciliation on changes.
			if err := ctrl.NewControllerManagedBy(mgr).
				For(&v1.Pod{}).
				Complete(reconciler); err != nil {
				return fmt.Errorf("failed to create controller: %w", err)
			}

			logger.Info("controller registered, starting manager")

			// Start controller manager.
			return mgr.Start(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")
	cmd.Flags().StringVar(&clusterName, "cluster", "", "Cluster name")

	return cmd
}
