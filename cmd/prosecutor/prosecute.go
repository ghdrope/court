package main

import (
	"time"

	"github.com/ghdrope/court/internal/prosecutor"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newProsecuteCommand starts the Prosecutor processing loop.
//
// The Prosecutor acts as an event-driven component that periodically
// fetches Incidentreports and analyzes them.
func newProsecuteCommand() *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "prosecute",
		Short: "Start Prosecutor event-processing loop",
		Long:  "Prosecutor continuously fetches and analyzes IncidentReports in an event-driven manner",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Initialize default service with mock implementations
			svc := prosecutor.NewService(
				prosecutor.NewMockFetcher(),
				prosecutor.NewStaticAnalyzer(),
			)

			zap.L().Info("prosecutor started",
				zap.Duration("interval", interval),
			)

			// Simulated processing loop
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					zap.L().Info("prosecutor shutting down")
					return nil

				case <-ticker.C:
					if err := svc.ProcessBatch(ctx); err != nil {
						zap.L().Error("processing error", zap.Error(err))
					}
				}
			}
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Polling interval for incident processing")

	return cmd
}
