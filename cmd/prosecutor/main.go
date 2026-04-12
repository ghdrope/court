package main

import (
	"os"

	"go.uber.org/zap"
	"k8s.io/sample-controller/pkg/signals"
)

func main() {
	// Context that is cancelled on SIGINT/SIGTERM
	ctx := signals.SetupSignalHandler()

	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// Ensure logger flushes buffered logs on exit
	defer func() {
		_ = logger.Sync()
	}()

	zap.ReplaceGlobals(logger)

	// Execute root command
	if err := Execute(ctx); err != nil {
		zap.L().Error("fatal error", zap.Error(err))
		os.Exit(1)
	}
}
