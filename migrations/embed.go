package migrations

import "embed"

// FS contains the database migration SQL scripts.
//
//go:embed *.sql
var FS embed.FS
