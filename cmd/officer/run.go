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

const defaultDsnAddr = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"
const defaultRedisAddr = "localhost:6379"

func runOfficer(
	ctx context.Context,
	redisAddr string,
	dsn string,
) error {

	logger := ctrl.Log.WithName("officer")
	logger.Info("starting officer")

	// ---------------------------
	// CONFIG
	// ---------------------------
	clusterName := env.Must("CLUSTER_NAME")

	databaseURL := env.FirstNonEmpty(dsn, env.Get("DATABASE_URL", defaultDsnAddr))
	redisAddress := env.FirstNonEmpty(redisAddr, env.Get("REDIS_ADDR", defaultRedisAddr))

	envMode := env.Get("ENV", "development")

	if envMode == "production" {
		if databaseURL == defaultDsnAddr {
			return fmt.Errorf("DATABASE_URL must be explicitly set in production")
		}
		if redisAddress == defaultRedisAddr {
			return fmt.Errorf("REDIS_ADDR must be explicitly set in production")
		}
	}

	logger.Info("configuration loaded",
		"cluster", clusterName,
		"env", envMode,
	)

	// ---------------------------
	// POSTGRES
	// ---------------------------
	db, err := postgres.Open(postgres.DefaultConfig(databaseURL))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(err, "failed to close db")
		}
	}()

	if err := db.PingWithRetry(ctx); err != nil {
		return fmt.Errorf("database not ready: %w", err)
	}

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
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	// ---------------------------
	// DOMAIN SERVICES
	// ---------------------------
	svc := officer.New(
		incidentRepo,
		suitRepo,
		rdb,
		logger,
	)

	// IMPORTANT: lifecycle manager (fix for your emitSuitCloseRequested issue)
	suitManager := &officer.SuitLifecycleManager{
		Log: logger.WithName("suit-lifecycle"),
		RDB: rdb,
	}

	// ---------------------------
	// K8S SCHEME + MANAGER
	// ---------------------------
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	config := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("create controller manager: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	// ---------------------------
	// RECONCILER
	// ---------------------------
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

	// register controller
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1.Pod{}).
		Complete(reconciler); err != nil {
		return fmt.Errorf("register controller: %w", err)
	}

	logger.Info("controller registered")

	// ---------------------------
	// RECOVERY PHASE (BEFORE LOOP START)
	// ---------------------------
	logger.Info("starting recovery phase")

	hints, err := svc.RecoverOpenIncidents(ctx, clusterName, kubeClient)
	if err != nil {
		return fmt.Errorf("recovery phase failed: %w", err)
	}

	// inject recovery hints into reconciler
	reconciler.RecoveryHints = hints

	logger.Info("recovery phase completed",
		"hints", len(hints),
	)

	// ---------------------------
	// START MANAGER LOOP
	// ---------------------------
	logger.Info("starting controller manager")

	return mgr.Start(ctx)
}
