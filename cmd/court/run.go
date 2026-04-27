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

	redisstream "github.com/ghdrope/court/internal/transport/redis"

	"github.com/ghdrope/court/pkg/env"
	"github.com/ghdrope/court/pkg/github"
	"github.com/ghdrope/court/pkg/postgres"
	redispkg "github.com/ghdrope/court/pkg/redis"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultRedisAddr = "localhost:6379"
	defaultDSN       = "postgres://postgres:postgres@localhost:5432/archive?sslmode=disable"
)

// runCourt bootstraps and starts the Court worker.
//
// It wires all dependencies and starts the event processing pipeline.
func runCourt(
	cmdCtx context.Context,
) error {

	logger := zap.L().With(zap.String("service", "court"))

	// --- Configuration ---
	databaseURL := env.Must("DATABASE_URL")
	redisAddress := env.Must("REDIS_ADDR")
	ghToken := env.Must("GITHUB_TOKEN")

	logger.Info("configuration loaded")

	// --- PostgreSQL ---
	db, err := postgres.Open(postgres.DefaultConfig(databaseURL))
	if err != nil {
		return fmt.Errorf("postgres open: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", zap.Error(err))
		}
	}()

	if err := db.PingWithRetry(cmdCtx); err != nil {
		return fmt.Errorf("postgres not ready: %w", err)
	}

	suitRepo := suit.NewRepository(db.DB)
	if err := suitRepo.InitSchema(cmdCtx); err != nil {
		return err
	}

	incidentRepo := incident.NewRepository(db.DB)
	if err := incidentRepo.InitSchema(cmdCtx); err != nil {
		return err
	}

	// --- Redis ---
	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddress,
	})

	hostname, _ := os.Hostname()
	consumerName := fmt.Sprintf("court-%s", hostname)

	incidentStreamClient := redispkg.NewStreamClient(
		rdb,
		redispkg.DefaultConfig(
			redisstream.IncidentCreatedStream,
			redisstream.CourtGroup,
			consumerName,
		),
	)

	closeStreamClient := redispkg.NewStreamClient(
		rdb,
		redispkg.DefaultConfig(
			redisstream.SuitCloseRequestedStream,
			redisstream.CourtGroup,
			consumerName,
		),
	)

	// --- VCS ---
	vcsClient := github.NewClient(ghToken)

	// --- Domain service ---
	svc := court.New(suitRepo, vcsClient, logger)

	// --- Consumers ---
	incidentConsumer := redisstream.NewIncidentCreatedConsumer(
		incidentStreamClient,
		incidentRepo,
		logger,
	)

	closeConsumer := redisstream.NewSuitCloseConsumer(
		closeStreamClient,
		logger,
	)

	logger.Info("court worker started",
		zap.String("redis", redisAddress),
		zap.String("consumer", consumerName),
	)

	// --- Run consumers concurrently ---
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

	// Block until context is cancelled
	<-cmdCtx.Done()

	logger.Info("court worker shutting down")

	return nil
}
