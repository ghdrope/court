package main

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the Prosecutor service.
// It acts as the entry point for all CLI commands.
var rootCmd = &cobra.Command{
	Use:   "prosecutor",
	Short: "Prosecutor service for Court analysis pipeline",
	Long:  "Prosecutor is an analysis service that evaluates incidents and places its comments.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute runs the CLI application with the provided context.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.AddCommand(newProsecuteCommand())
}
