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
