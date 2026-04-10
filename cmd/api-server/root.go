package main

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the API server.
var rootCmd = &cobra.Command{
	Use:   "api-server",
	Short: "Stateless routing API server",
	Long:  "A gRPC API server that routes IncidentReports between Officer and Court services.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute runs the CLI.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.AddCommand(newServeCommand())
}
