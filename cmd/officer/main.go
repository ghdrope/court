package main

import (
	"os"

	"go.uber.org/zap"
	"k8s.io/sample-controller/pkg/signals"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	// Setup a context that is automatically cancelled on SIGINT/SIGTERM.
	ctx := signals.SetupSignalHandler()

	// Initialize global logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctrl.SetLogger(ctrlzap.New(ctrlzap.UseDevMode(false)))

	// Execute CLI
	if err := Execute(ctx); err != nil {
		zap.L().Error("fatal error", zap.Error(err))
		os.Exit(1)
	}
}
