package dbstart

import (
	"context"
	"database/sql"
)

// SchemaVersion returns the current schema version from the database.
func SchemaVersion(ctx context.Context, db *sql.DB) (int, error) {
	return ensureVersionTable(ctx, db)
}
