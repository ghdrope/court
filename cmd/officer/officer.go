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
	"go.uber.org/zap"

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

	var (
		redisAddr string
		cluster   string
		dsn       string
	)

	cmd := &cobra.Command{
		Use:  "officer",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {

			ctrlLogger := ctrl.Log.WithName("officer")
			ctrlLogger.Info("starting officer")

			// Resolve config (flag > env > default)
			clusterName := env.FirstNonEmpty(
				cluster,
				env.Get("CLUSTER_NAME", defaultClusterName),
			)

			databaseURL := env.FirstNonEmpty(
				dsn,
				env.Get("DATABASE_URL", defaultDsnAddr),
			)

			redisAddress := env.FirstNonEmpty(
				redisAddr,
				env.Get("REDIS_ADDR", defaultRedisAddr),
			)

			// Initialize structured logger (zap).
			zapLogger := zap.L().With(zap.String("component", "officer"))

			// --- Postgres ---
			// Initialize PostgreSQL connection.
			// This is required for persisting incident storage.
			db, err := postgres.Open(databaseURL)
			if err != nil {
				return fmt.Errorf("db open: %w", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					zapLogger.Error("failed to close database", zap.Error(err))
				}
			}()

			// Ensure database readiness before controller startup.
			if err := postgres.PingWithRetry(cmd.Context(), db); err != nil {
				return fmt.Errorf("db not ready: %w", err)
			}

			// Initialize incident repository layer.
			repo := incident.NewRepository(db)

			// Ensure database schema exists before processing workloads.
			if err := repo.InitSchema(cmd.Context()); err != nil {
				return fmt.Errorf("init schema: %w", err)
			}

			// --- Redis ---
			// Initialize Redis client for event publishing.
			rdb := redis.NewClient(&redis.Options{
				Addr: redisAddress,
			})

			// --- Service ---
			// Initialize Officer service.
			svc := officer.New(repo, rdb, zapLogger)

			// --- Kubernetes ---
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
			kubeClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}

			// Initialize Pod reconciler.
			// It detects failures and generates IncidentReports.
			reconciler := &officer.PodReconciler{
				Client:     mgr.GetClient(),
				KubeClient: kubeClient,
				Log:        log.Log.WithName("reconciler"),
				Service:    svc,
				Cluster:    clusterName,
			}

			// Register controller with manager.
			// Watches Pod resources and triggers reconciliation on changes.
			if err := ctrl.NewControllerManagedBy(mgr).
				For(&v1.Pod{}).
				Complete(reconciler); err != nil {
				return fmt.Errorf("failed to create controller: %w", err)
			}

			ctrlLogger.Info("controller registered, starting manager")

			// Start controller manager.
			return mgr.Start(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")
	cmd.Flags().StringVar(&dsn, "database-url", "", "PostgreSQL DSN")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")

	return cmd
}
