package main

import (
	"context"
	"fmt"

	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/officer"
	"github.com/ghdrope/court/internal/suit"
	"github.com/ghdrope/court/pkg/env"
	"github.com/ghdrope/court/pkg/postgres"

	"github.com/redis/go-redis/v9"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// runOfficer is the main application for the Officer runtime.
//
// It wires together all dependencies:
//
//   - configuration (env vars)
//   - persistence (PostgreSQL)
//   - event bus (Redis)
//   - K8s client + controller manager
//   - services and reconcilers
//
// The function is long-lived and blocks until the
// controller manager exits (via context cancellation).
func runOfficer(ctx context.Context) error {

	logger := ctrl.Log.WithName("officer")
	logger.Info("starting officer")

	// ---------------------------
	// CONFIGURATION
	// ---------------------------
	// All configuration is strictly required at startup.
	// Missing values cause immediate failure.
	databaseURL := env.Must("DATABASE_URL")
	redisAddress := env.Must("REDIS_ADDR")
	clusterName := env.Must("CLUSTER_NAME")

	logger.Info("configuration loaded",
		"cluster", clusterName,
	)

	// ---------------------------
	// POSTGRESQL
	// ---------------------------
	// PostgreSQL is the source of persistence for incidents and suits.
	db, err := postgres.Open(postgres.DefaultConfig(databaseURL))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(err, "failed to close database")
		}
	}()

	// Ensure database is reachable before starting controllers
	if err := db.PingWithRetry(ctx); err != nil {
		return fmt.Errorf("database not ready: %w", err)
	}

	// Initialize persistence schemas
	incidentRepo := incident.NewRepository(db.DB)
	if err := incidentRepo.InitSchema(ctx); err != nil {
		return fmt.Errorf("init incident schema: %w", err)
	}

	suitRepo := suit.NewRepository(db.DB)
	if err := suitRepo.InitSchema(ctx); err != nil {
		return fmt.Errorf("init suit schema: %w", err)
	}

	// ---------------------------
	// REDIS
	// ---------------------------
	// Redis is used for lifecycle events signals.
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	// ---------------------------
	// SERVICES
	// ---------------------------
	// Officer service
	svc := officer.New(
		incidentRepo,
		suitRepo,
		rdb,
		logger,
	)

	// Manages lifecycle transitions for Suits.
	suitManager := &officer.SuitLifecycleManager{
		Log: logger.WithName("suit-lifecycle"),
		RDB: rdb,
	}

	// ---------------------------
	// K8S RUNTIME
	// ---------------------------
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	config := ctrl.GetConfigOrDie()

	// Controller manager is the runtime orchestrator for reconcilers.
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("create controller manager: %w", err)
	}
	logger.Info("controller manager initialized")

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	// ---------------------------
	// RECONCILER
	// ---------------------------
	// PodReconciler translates K8s state into events.
	reconciler := &officer.PodReconciler{
		Client:      mgr.GetClient(),
		KubeClient:  kubeClient,
		Log:         log.Log.WithName("reconciler"),
		Service:     svc,
		SuitRepo:    suitRepo,
		SuitManager: suitManager,
		Cluster:     clusterName,
		RDB:         rdb,
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1.Pod{}).
		Complete(reconciler); err != nil {
		return fmt.Errorf("register controller: %w", err)
	}

	logger.Info("controller registered")

	// ---------------------------
	// RECOVERY PHASE
	// ---------------------------
	// Before starting reconciliation loop, restore state.
	// to avoid inconsistency after restarts.
	logger.Info("starting recovery phase")

	hints, err := svc.RecoverOpenIncidents(ctx, clusterName, kubeClient)
	if err != nil {
		return fmt.Errorf("recovery phase failed: %w", err)
	}

	reconciler.RecoveryHints = hints

	logger.Info("recovery phase completed",
		"hints", len(hints),
	)

	// ---------------------------
	// START CONTROLLER LOOP
	// ---------------------------
	logger.Info("starting controller manager loop")

	// Blocks 'til context cancellation
	return mgr.Start(ctx)
}
