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

	"go.uber.org/zap"
	"k8s.io/sample-controller/pkg/signals"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// main is the process entrypoint.
//
// It initializes structured logging, installs signal handlers for graceful
// shutdown (SIGINT/SIGTERM), and executes the CLI root command.
//
// Any fatal error during execution results in a non-zero exit code.
func main() {
	ctx := signals.SetupSignalHandler()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Replace global zap logger
	zap.ReplaceGlobals(logger)

	// Configure controller-runtime logger
	ctrl.SetLogger(ctrlzap.New(ctrlzap.UseDevMode(false)))

	if err := Execute(ctx); err != nil {
		zap.L().Error("fatal error", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("shutdown complete")
}
