package main

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the Archive.
var rootCmd = &cobra.Command{
	Use:   "archive",
	Short: "",
	Long:  "",
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
	rootCmd.AddCommand(newArchiveCommand())
}
