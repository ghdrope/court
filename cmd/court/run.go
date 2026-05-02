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
	"context"
	"fmt"
	"os"

	"github.com/ghdrope/court/internal/court"
	"github.com/ghdrope/court/internal/incident"
	"github.com/ghdrope/court/internal/suit"

	docket "github.com/ghdrope/court/internal/docket"

	"github.com/ghdrope/court/pkg/env"
	"github.com/ghdrope/court/pkg/github"
	"github.com/ghdrope/court/pkg/postgres"
	redispkg "github.com/ghdrope/court/pkg/redis"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// runCourt bootstraps and runs the Court worker process.
//
// It wires together all dependencies:
//
//   - configuration (env vars)
//   - persistence (PostgreSQL)
//   - event bus (Redis)
//   - external API clients (integrated VCSs)
//   - services and event bus consumers
//
// The function blocks until the provided context is cancelled.
func runCourt(cmdCtx context.Context) error {

	logger := zap.L().With(zap.String("service", "court"))

	// ---------------------------
	// CONFIGURATION
	// ---------------------------
	// All configuration is strictly required at startup.
	// Missing values cause immediate failure.
	databaseURL := env.Must("DATABASE_URL")
	redisAddress := env.Must("REDIS_ADDR")
	ghToken := env.Must("GITHUB_TOKEN")

	logger.Info("configuration loaded")

	// ---------------------------
	// POSTGRESQL
	// ---------------------------
	// PostgreSQL is the source of persistence for suits.
	db, err := postgres.Open(postgres.DefaultConfig(databaseURL))
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", zap.Error(err))
		}
	}()

	// Ensure database is reachable before starting controllers
	if err := db.PingWithRetry(cmdCtx); err != nil {
		return fmt.Errorf("database not ready: %w", err)
	}

	// Initialize persistence schemas
	suitRepo := suit.NewRepository(db.DB)
	if err := suitRepo.InitSchema(cmdCtx); err != nil {
		return fmt.Errorf("init incident schema: %w", err)
	}

	incidentRepo := incident.NewRepository(db.DB)
	if err := incidentRepo.InitSchema(cmdCtx); err != nil {
		return fmt.Errorf("init suit schema: %w", err)
	}

	// ---------------------------
	// REDIS
	// ---------------------------
	// Redis is used for lifecycle events signals consume.
	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddress,
	})

	hostname, _ := os.Hostname()
	consumerName := fmt.Sprintf("court-%s", hostname)

	// Event bus: incident creation events
	incidentStreamClient := redispkg.NewStreamClient(
		rdb,
		redispkg.DefaultConfig(
			docket.IncidentCreatedStream,
			docket.CourtGroup,
			consumerName,
		),
	)

	// Event bus: suit close requests
	closeStreamClient := redispkg.NewStreamClient(
		rdb,
		redispkg.DefaultConfig(
			docket.SuitCloseRequestedStream,
			docket.CourtGroup,
			consumerName,
		),
	)

	// ---------------------------
	// VCS EXTERNAL CLIENTS
	// ---------------------------
	// GitHub
	vcsClient := github.NewClient(ghToken)

	// ---------------------------
	// SERVICES
	// ---------------------------
	// Court service
	svc := court.New(
		suitRepo,
		vcsClient,
		logger,
	)

	// ---------------------------
	// CONSUMERS
	// ---------------------------
	// Each consumer runs independently and processes a single event stream.

	incidentConsumer := docket.NewIncidentCreatedConsumer(
		incidentStreamClient,
		incidentRepo,
		logger,
	)

	closeConsumer := docket.NewSuitCloseConsumer(
		closeStreamClient,
		logger,
	)

	logger.Info("court worker started",
		zap.String("redis", redisAddress),
		zap.String("consumer", consumerName),
	)

	// ---------------------------
	// CONCURRENT EXECUTION
	// ---------------------------
	// Consumers run independently and are expected to run indefinitely
	// until the context is cancelled.
	//
	// Any fatal error inside a consumer terminates the process.
	go func() {
		if err := incidentConsumer.Start(cmdCtx, svc); err != nil {
			logger.Fatal("incident consumer failed", zap.Error(err))
		}
	}()

	go func() {
		if err := closeConsumer.Start(cmdCtx, svc); err != nil {
			logger.Fatal("close consumer failed", zap.Error(err))
		}
	}()

	// // Blocks 'til context cancellation
	<-cmdCtx.Done()

	return nil
}
