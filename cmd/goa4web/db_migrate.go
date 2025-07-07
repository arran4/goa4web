package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"os"

	dbpkg "github.com/arran4/goa4web/internal/db"
	dbdrivers "github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/migrate"
	"github.com/arran4/goa4web/runtimeconfig"
)

// openDB establishes a database connection without verifying the schema version.
func openDB(cfg runtimeconfig.RuntimeConfig) (*sql.DB, error) {
	conn := cfg.DBConn
	if conn == "" {
		return nil, fmt.Errorf("connection string required")
	}
	c, err := dbdrivers.Connector(cfg.DBDriver, conn)
	if err != nil {
		return nil, err
	}
	var connector driver.Connector = dbpkg.NewLoggingConnector(c)
	db := sql.OpenDB(connector)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// dbMigrateCmd implements "db migrate".
type dbMigrateCmd struct {
	*dbCmd
	fs   *flag.FlagSet
	Dir  string
	args []string
}

func parseDbMigrateCmd(parent *dbCmd, args []string) (*dbMigrateCmd, error) {
	c := &dbMigrateCmd{dbCmd: parent}
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.StringVar(&c.Dir, "dir", "migrations", "directory containing SQL migrations")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *dbMigrateCmd) Run() error {
	if c.rootCmd.Verbosity >= 0 {
		fmt.Printf("connecting to database at %s:%s\n", c.rootCmd.cfg.DBHost,
			c.rootCmd.cfg.DBPort)
	}
	db, err := openDB(c.rootCmd.cfg)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()
	ctx := context.Background()
	fsys := os.DirFS(c.Dir)
	if c.rootCmd.Verbosity >= 0 {
		fmt.Printf("applying migrations from %s\n", c.Dir)
	}
	if err := migrate.Apply(ctx, db, fsys, c.rootCmd.Verbosity >= 0); err != nil {
		return err
	}
	return nil
}
