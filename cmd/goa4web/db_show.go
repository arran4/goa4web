package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/migrations"
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
	case "schema", "schema.mysql.sql":
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
		if c.target == "migrations" {
			entries, err := fs.ReadDir(migrations.FS, ".")
			if err != nil {
				return err
			}
			for _, e := range entries {
				fmt.Println(e.Name())
			}
			return nil
		}

		b, err := fs.ReadFile(migrations.FS, c.target)
		if err == nil {
			_, err = fmt.Fprintln(os.Stdout, string(b))
			return err
		}

		return fmt.Errorf("unknown target %q", c.target)
	}
}
