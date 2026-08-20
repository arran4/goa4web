package dbstart

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// SchemaVersion returns the current schema version from the database using the default mysql dialect.
func SchemaVersion(ctx context.Context, db *sql.DB) (int, error) {
	return SchemaVersionWithDriver(ctx, db, "mysql")
}

// SchemaVersionWithDriver returns the current schema version from the database for the given driver.
func SchemaVersionWithDriver(ctx context.Context, db *sql.DB, driver string) (int, error) {
	dialect := "mysql"
	if driver == "sqlite" || driver == "sqlite3" {
		dialect = "sqlite3"
	}
	if err := goose.SetDialect(dialect); err != nil {
		return 0, err
	}
	version, err := goose.GetDBVersion(db)
	return int(version), err
}
