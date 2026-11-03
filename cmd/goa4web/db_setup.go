package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/sqlutil"
	"github.com/arran4/goa4web/migrations"
)

// dbSetupCmd implements "db setup".
type dbSetupCmd struct {
	*dbCmd
}

func (c *dbSetupCmd) FlagGroups() []flagGroup {
	return nil
}

var _ usageData = (*dbSetupCmd)(nil)

// Usage prints command usage.
func (c *dbSetupCmd) Usage() {
	_ = executeUsage(c.rootCmd.fs.Output(), "db_setup_usage.txt", c)
}

func parseDbSetupCmd(parent *dbCmd, args []string) (*dbSetupCmd, error) {
	c := &dbSetupCmd{dbCmd: parent}
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}
	return c, nil
}

func (c *dbSetupCmd) Run() error {
	cfg := c.cfg
	conn := cfg.DBConn
	if conn == "" {
		return fmt.Errorf("connection string required")
	}
	connector, err := c.dbReg.Connector(cfg.DBDriver, conn)
	if err != nil {
		return err
	}
	sdb := sql.OpenDB(connector)
	defer func(sdb *sql.DB) {
		err := sdb.Close()
		if err != nil {
			log.Printf("failed to close db connection: %v", err)
		}
	}(sdb)
	if err := sdb.Ping(); err != nil {
		return err
	}

	ctx := context.Background()
	if cfg.DBDriver == "sqlite" || cfg.DBDriver == "sqlite3" {
		log.Println("Applying SQLite migrations...")
		if err := dbstart.Apply(ctx, sdb, migrations.FS, false, cfg.DBDriver); err != nil {
			return fmt.Errorf("apply SQLite migrations: %w", err)
		}
	} else {
		log.Println("Applying schema...")
		if err := sqlutil.RunStatements(ctx, sdb, strings.NewReader(string(database.SchemaMySQL))); err != nil {
			return fmt.Errorf("failed to apply schema: %w", err)
		}
	}

	log.Println("Applying seed data...")
	if err := sqlutil.RunStatements(ctx, sdb, strings.NewReader(string(database.SeedSQLForDriver(cfg.DBDriver)))); err != nil {
		return fmt.Errorf("failed to apply seed data: %w", err)
	}
	if cfg.DBDriver == "sqlite" || cfg.DBDriver == "sqlite3" {
		if _, err := sdb.ExecContext(ctx, "INSERT INTO language (id, nameof) VALUES (1, 'English') ON CONFLICT DO NOTHING"); err != nil {
			return fmt.Errorf("seed default language: %w", err)
		}
	}

	log.Println("Database setup successfully.")
	return nil
}
