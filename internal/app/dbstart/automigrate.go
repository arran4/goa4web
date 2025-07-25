package dbstart

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
)

// autoMigrateEnabled reports whether automatic migrations should run.
func autoMigrateEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvAutoMigrate))
	switch v {
	case "1", "true", "on", "yes":
		return true
	default:
		return false
	}
}

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
	var connector driver.Connector = dbpkg.NewLoggingConnector(c, cfg.DBLogVerbosity)
	db := sql.OpenDB(connector)
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	fsys := os.DirFS("migrations")
	return Apply(ctx, db, fsys, false)
}

// MaybeAutoMigrate runs migrations when enabled via AUTO_MIGRATE.
func MaybeAutoMigrate(cfg *config.RuntimeConfig, reg *dbdrivers.Registry) error {
	if !autoMigrateEnabled() {
		return nil
	}
	ctx := context.Background()
	if err := applyMigrations(ctx, cfg, reg); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	return nil
}
