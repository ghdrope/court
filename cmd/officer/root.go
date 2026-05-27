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

// rootCmd defines the base CLI endpoint.
//
// It exists to:
//   - provide CLI structure
//   - enable context propagation across Cobra command tree
//   - display help when invoked without arguments
var rootCmd = &cobra.Command{
	Use:   "officer",
	Short: "Court Officer CLI",
	Long:  "Officer monitors Kubernetes workloads and reports incidents.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		// Default behavior when no subcommand is provided
		_ = cmd.HelpFunc()
	},
}

// Execute runs the CLI command tree using the provided context.
//
// The context is propagated to all subcommands and long-running
// operations.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// Register runtime subcommands
	rootCmd.AddCommand(newOfficerCommand())
	rootCmd.AddCommand(newVersionCommand())
}
