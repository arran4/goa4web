package migrations

import "embed"

// FS contains the database migration SQL scripts. Each new migration must
// update the `schema_version` table and bump
// handlers.ExpectedSchemaVersion.
//
//go:embed *.sql
var FS embed.FS
