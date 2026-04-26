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
	"os"

	"github.com/spf13/cobra"
)

// newOfficerCommand creates the runtime responsible for
// starting the Officer controller.
func newOfficerCommand() *cobra.Command {

	var (
		redisAddr string
		dsn       string
		envMode   string
	)

	cmd := &cobra.Command{
		Use:   "officer",
		Short: "Start the Officer controller",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// CLI flag overrides environment variable
			if envMode != "" {
				_ = os.Setenv("ENV", envMode)
			}
			return runOfficer(
				cmd.Context(),
				redisAddr,
				dsn,
			)
		},
	}

	cmd.Flags().StringVar(&redisAddr, "redis-addr", "", "Redis address")
	cmd.Flags().StringVar(&dsn, "database-url", "", "PostgreSQL DSN")
	cmd.Flags().StringVar(&envMode, "env", "", "Runtime environment (development|production)")

	return cmd
}
