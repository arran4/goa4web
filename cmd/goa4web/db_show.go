package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/app/dbstart"
)

// dbShowCmd implements "db show".
type dbShowCmd struct {
	*dbCmd
	target string // The file to show, e.g., "seed.sql"
}

func (c *dbShowCmd) FlagGroups() []flagGroup {
	return nil
}

var _ usageData = (*dbShowCmd)(nil)

// Usage prints command usage.
func (c *dbShowCmd) Usage() {
	executeUsage(c.rootCmd.fs.Output(), "db_show_usage.txt", c)
}

func parseDbShowCmd(parent *dbCmd, args []string) (*dbShowCmd, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected exactly one argument (e.g., 'seed.sql' or 'schema.mysql.sql'), but got %d", len(args))
	}
	c := &dbShowCmd{
		dbCmd:  parent,
		target: args[0],
	}
	return c, nil
}

func (c *dbShowCmd) Run() error {
	switch strings.ToLower(c.target) {
	case "seed.sql":
		_, err := fmt.Fprintln(os.Stdout, string(database.SeedSQL))
		return err
	case "schema.mysql.sql":
		_, err := fmt.Fprintln(os.Stdout, string(database.SchemaMySQL))
		return err
	case "schema-version":
		db, err := openDB(c.rootCmd.cfg, c.rootCmd.dbReg)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer db.Close()
		version, err := dbstart.SchemaVersion(context.Background(), db)
		if err != nil {
			return fmt.Errorf("read schema version: %w", err)
		}
		_, err = fmt.Fprintln(os.Stdout, version)
		return err
	default:
		return fmt.Errorf("unknown target %q", c.target)
	}
}
