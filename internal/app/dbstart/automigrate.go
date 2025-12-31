package dbstart

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/arran4/goa4web/migrations"
	"io/fs"
	"log"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
)

// applyMigrations connects to the database and executes SQL migrations.
func applyMigrations(ctx context.Context, cfg *config.RuntimeConfig, reg *dbdrivers.Registry) error {
	conn := cfg.DBConn
	if conn == "" {
		return fmt.Errorf("connection string required")
	}
	c, err := reg.Connector(cfg.DBDriver, conn)
	if err != nil {
		return err
	}
	var connector driver.Connector = db.NewLoggingConnector(c, cfg.DBLogVerbosity)
	sdb := sql.OpenDB(connector)
	defer func(sdb *sql.DB) {
		err := sdb.Close()
		if err != nil {
			log.Printf("failed to close DB connection: %v", err)
		}
	}(sdb)
	if err := sdb.PingContext(ctx); err != nil {
		return err
	}
	var fsys fs.FS
	if cfg.MigrationsDir == "" || cfg.MigrationsDir == "migrations" {
		fsys = migrations.FS
	} else {
		fsys = os.DirFS(cfg.MigrationsDir)
	}

	current, err := ensureVersionTable(ctx, sdb)
	if err != nil {
		return fmt.Errorf("check version: %w", err)
	}

	found, err := getAvailableMigrations(fsys, cfg.DBDriver)
	if err != nil {
		return fmt.Errorf("check migrations: %w", err)
	}

	target := current
	if len(found) > 0 {
		max := found[len(found)-1]
		if max.Version > target {
			target = max.Version
		}
	}

	if cfg.MigrationsDir == "" || cfg.MigrationsDir == "migrations" {
		log.Printf("applying embedded migrations (current: %d, target: %d)", current, target)
	} else {
		log.Printf("applying migrations from %s (current: %d, target: %d)", cfg.MigrationsDir, current, target)
	}

	return Apply(ctx, sdb, fsys, false, cfg.DBDriver)
}

// MaybeAutoMigrate runs migrations when enabled via runtime configuration.
func MaybeAutoMigrate(cfg *config.RuntimeConfig, reg *dbdrivers.Registry) error {
	if !cfg.AutoMigrate {
		return nil
	}
	ctx := context.Background()
	if err := applyMigrations(ctx, cfg, reg); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	return nil
}
