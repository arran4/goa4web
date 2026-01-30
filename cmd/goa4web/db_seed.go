package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/sqlutil"
)

// dbSeedCmd implements "db seed".
type dbSeedCmd struct {
	*dbCmd
}

func (c *dbSeedCmd) FlagGroups() []flagGroup {
	return nil
}

var _ usageData = (*dbSeedCmd)(nil)

// Usage prints command usage.
func (c *dbSeedCmd) Usage() {
	executeUsage(c.rootCmd.fs.Output(), "db_seed_usage.txt", c)
}

func parseDbSeedCmd(parent *dbCmd, args []string) (*dbSeedCmd, error) {
	c := &dbSeedCmd{dbCmd: parent}
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}
	return c, nil
}

func (c *dbSeedCmd) Run() error {
	cfg := c.rootCmd.cfg
	conn := cfg.DBConn
	if conn == "" {
		return fmt.Errorf("connection string required")
	}
	connector, err := c.rootCmd.dbReg.Connector(cfg.DBDriver, conn)
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

	log.Println("Applying seed data...")
	if err := sqlutil.RunStatements(context.Background(), sdb, strings.NewReader(string(database.SeedSQL))); err != nil {
		return fmt.Errorf("failed to apply seed data: %w", err)
	}

	log.Println("Database seeded successfully.")
	return nil
}
