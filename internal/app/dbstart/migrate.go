package dbstart

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"

	"github.com/arran4/goa4web/migrations"
	"github.com/pressly/goose/v3"
)

func legacySchemaVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	err := db.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		// Table might not exist
		return 0, nil
	}
	return version, nil
}

func transitionToGoose(ctx context.Context, db *sql.DB) error {
	legacyVer, err := legacySchemaVersion(ctx, db)
	if err != nil {
		return err
	}
	if legacyVer == 0 {
		return nil
	}

	gooseVer, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("get goose db version: %w", err)
	}

	if gooseVer == 0 {
		// Sync Goose with the legacy version
		for i := 1; i <= legacyVer; i++ {
			if _, err := db.ExecContext(ctx, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)", i); err != nil {
				return fmt.Errorf("sync goose_db_version: %w", err)
			}
		}
		log.Printf("transitioned to goose_db_version at version %d", legacyVer)
	}
	return nil
}

// Apply runs the migrations using Goose.
func Apply(ctx context.Context, db *sql.DB, f fs.FS, verbose bool, driver string) error {
	goose.SetBaseFS(migrations.FilterFS(f, driver))
	dialect := driver
	if dialect == "sqlite" {
		dialect = "sqlite3"
	}
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if _, err := goose.EnsureDBVersion(db); err != nil {
		return fmt.Errorf("ensure goose db version: %w", err)
	}
	if err := transitionToGoose(ctx, db); err != nil {
		return fmt.Errorf("transition to goose: %w", err)
	}
	// Enable goose verbose logging if requested, but there's no direct SetVerbose in goose v3
	// We just run UpContext
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
