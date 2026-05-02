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

package postgres

import "strings"

// SanitizeDSN redacts sensitive credentials from a PostgreSQL DSN.
//
// It removes the username and password portion between "://"
// and "@" to avoid leaking credentials in logs.
//
// Example:
//
// postgres://user:password@host:5432/db
// -> postgres://***:***@host:5432/db
func SanitizeDSN(dsn string) string {
	return redactBetween(dsn, "://", "@")
}

// redactBetween replaces the substring between start and end markers
// with a fixed redacted value.
//
// If markers are not found, the original string is returned unchanged.
func redactBetween(s, start, end string) string {
	i := strings.Index(s, start)
	j := strings.Index(s, end)

	if i == -1 || j == -1 || j <= i {
		return s
	}

	return s[:i+len(start)] + "***:***" + s[j:]
}
