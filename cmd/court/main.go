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
	"os"

	"github.com/ghdrope/court/pkg/utils"
	"go.uber.org/zap"
	"k8s.io/sample-controller/pkg/signals"
)

// main bootstraps the Court process.
//
// Responsibilities:
//   - initialize logging
//   - install OS signal handling for graceful shutdown
//   - delegate execution to the CLI root command
//
// The process exits with status 1 if any fatal error occurs during runtime.
func main() {
	// ctx is cancelled on SIGINT/SIGTERM to ensure graceful shutdown
	ctx := signals.SetupSignalHandler()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Set global logger
	zap.ReplaceGlobals(logger)

	if utils.IsDebug() {
		zap.L().Info("🐛 DEBUG MODE ENABLED")
	}

	if err := Execute(ctx); err != nil {
		zap.L().Error("fatal error during execution", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("shutdown complete")
}
