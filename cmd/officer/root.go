package main

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd is the Officer base CLI command.
var rootCmd = &cobra.Command{
	Use:   "officer",
	Short: "Court Officer CLI",
	Long:  `CLI for monitoring Kubernetes pod health.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute runs the CLI with context propagation.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

// init registers subcommands.
func init() {
	rootCmd.AddCommand(newPatrolCommand())
}
