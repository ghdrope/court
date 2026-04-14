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

// rootCmd is the base CLI command for the Prosecutor.
var rootCmd = &cobra.Command{
	Use:   "prosecutor",
	Short: "Prosecutor service for Court analysis pipeline",
	Long:  "Prosecutor is an analysis service that looksup to enrich data stored at the Archive.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute runs the CLI with context propagation.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.AddCommand(newProsecutorCommand())
}
