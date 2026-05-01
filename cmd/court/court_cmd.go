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

import "github.com/spf13/cobra"

// newCourtCommand defines the runtime entrypoint for starting
// the Court process.
//
// This command is responsible for:
//   - connecting to external systems (DB, Redis)
//   - initializing services
//   - starting the event bus consumer loop
func newCourtCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "court",
		Short: "Start Court worker",

		RunE: func(cmd *cobra.Command, args []string) error {

			// Attribute execution to application runtime
			return runCourt(
				cmd.Context(),
			)
		},
	}

	return cmd
}
