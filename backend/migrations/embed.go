// Package migrations holds embedded goose SQL files for runtime migration (see internal/adapters/secondary/database/postgres.go).
// CLI: make migrate-up (runs goose against ./migrations on disk — same files).
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS
