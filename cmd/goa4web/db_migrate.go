package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"flag"
	"fmt"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/migrations"
	"io/fs"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
)

// openDB establishes a database connection without verifying the schema version.
func openDB(cfg *config.RuntimeConfig, reg *dbdrivers.Registry) (*sql.DB, error) {
	conn := cfg.DBConn
	if conn == "" {
		return nil, fmt.Errorf("connection string required")
	}
	c, err := reg.Connector(cfg.DBDriver, conn)
	if err != nil {
		return nil, err
	}
	var connector driver.Connector = db.NewLoggingConnector(c, cfg.DBLogVerbosity)
	sdb := sql.OpenDB(connector)
	if err := sdb.Ping(); err != nil {
		return nil, closeAndWrap(sdb, fmt.Errorf("ping database: %w", err))
	}
	return sdb, nil
}

func closeAndWrap(db *sql.DB, err error) error {
	if cerr := db.Close(); cerr != nil {
		return errors.Join(err, fmt.Errorf("close db: %w", cerr))
	}
	return err
}

// dbMigrateCmd implements "db migrate".
type dbMigrateCmd struct {
	*dbCmd
	fs  *flag.FlagSet
	Dir string
}

func parseDbMigrateCmd(parent *dbCmd, args []string) (*dbMigrateCmd, error) {
	c := &dbMigrateCmd{dbCmd: parent}
	c.fs = newFlagSet("migrate")
	c.fs.StringVar(&c.Dir, "dir", "migrations", "directory containing SQL migrations")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *dbMigrateCmd) Usage() {
	executeUsage(c.fs.Output(), "db_migrate_usage.txt", c)
}

func (c *dbMigrateCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*dbMigrateCmd)(nil)

func (c *dbMigrateCmd) Run() error {
	c.rootCmd.Verbosef("connecting to database using %s", c.rootCmd.cfg.DBConn)
	db, err := openDB(c.rootCmd.cfg, c.rootCmd.dbReg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()
	ctx := context.Background()
	var fsys fs.FS
	if c.Dir == "migrations" {
		fsys = migrations.FS
		c.rootCmd.Verbosef("applying embedded migrations")
	} else {
		fsys = os.DirFS(c.Dir)
		c.rootCmd.Verbosef("applying migrations from %s", c.Dir)
	}
	if err := dbstart.Apply(ctx, db, fsys, c.rootCmd.Verbosity >= 0, c.rootCmd.cfg.DBDriver); err != nil {
		return err
	}
	c.rootCmd.Infof("database migrated successfully")
	return nil
}
