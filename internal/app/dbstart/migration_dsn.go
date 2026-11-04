package dbstart

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	sqlmysql "github.com/go-sql-driver/mysql"
)

// MigrationConnectionString selects the user-supplied migration connection or,
// if omitted, the application connection. Neither DSN is modified.
func MigrationConnectionString(cfg *config.RuntimeConfig) string {
	if cfg.DBMigrationConn != "" {
		return cfg.DBMigrationConn
	}
	return cfg.DBConn
}

// ValidateMigrationConnection checks that pending MySQL migrations can execute
// Goose multi-statement blocks. Already-current schemas do not require this
// option just to start the application.
func ValidateMigrationConnection(ctx context.Context, db *sql.DB, driverName, dsn string) error {
	if driverName != "mysql" {
		return nil
	}
	cfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("parse MySQL migration connection: %w", err)
	}
	if cfg.MultiStatements {
		return nil
	}
	legacyVersion, err := legacySchemaVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("inspect schema version before migration: %w", err)
	}
	if legacyVersion >= handlers.ExpectedSchemaVersion {
		return nil
	}
	return fmt.Errorf("schema migration from version %d to %d requires multiStatements=true on the MySQL migration connection; configure DB_MIGRATION_CONN (preferred) or DB_CONN explicitly; Goa4Web will not modify either DSN", legacyVersion, handlers.ExpectedSchemaVersion)
}
