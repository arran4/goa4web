package dbstart

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// SchemaVersion returns the current schema version from the database.
func SchemaVersion(ctx context.Context, db *sql.DB) (int, error) {
	goose.SetDialect("mysql")
	version, err := goose.GetDBVersion(db)
	return int(version), err
}
