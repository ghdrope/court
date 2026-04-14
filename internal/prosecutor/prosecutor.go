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

package prosecutor

import "database/sql"

// Service handles post-processing over items previously persisted in
// the Archive.
//
// It is responsible for enriching stored data.
type Service struct {
	DB *sql.DB
}

// New creates a new Prosecutor service backed by the given database.
func New(db *sql.DB) *Service {
	return &Service{DB: db}
}
