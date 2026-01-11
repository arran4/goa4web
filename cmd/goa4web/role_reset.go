package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// roleResetCmd implements the "role reset" subcommand.
type roleResetCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	role string
	file string
}

func parseRoleResetCmd(parent *roleCmd, args []string) (*roleResetCmd, error) {
	c := &roleResetCmd{roleCmd: parent}
	fs := flag.NewFlagSet("reset", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.role, "role", "", "The name of the role to reset.")
	fs.StringVar(&c.file, "file", "", "Optional path to a .sql file to load instead of the embedded role script.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.role == "" {
		return nil, fmt.Errorf("role name is required")
	}
	return c, nil
}

func (c *roleResetCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.New(sdb)
	ctx := c.rootCmd.ctx

	role, err := q.GetRoleByName(ctx, c.role)
	if err != nil {
		return fmt.Errorf("failed to get role by name: %w", err)
	}

	log.Printf("Deleting grants for role %q (ID: %d)", c.role, role.ID)
	if err := q.DeleteGrantsByRoleID(ctx, sql.NullInt32{Int32: role.ID, Valid: true}); err != nil {
		return fmt.Errorf("failed to delete grants: %w", err)
	}

	var data []byte
	if c.file != "" {
		p := c.file
		if !strings.HasSuffix(strings.ToLower(p), ".sql") {
			p = p + ".sql"
		}
		abs, _ := filepath.Abs(p)
		log.Printf("Loading role %q from file %s", c.role, abs)
		b, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("failed to read role file: %w", err)
		}
		data = b
	} else {
		log.Printf("Loading role %q from embedded roles", c.role)
		b, err := readEmbeddedRole(c.role)
		if err != nil {
			return fmt.Errorf("failed to read embedded role %q: %w", c.role, err)
		}
		data = b
	}

	if err := runStatements(sdb, strings.NewReader(string(data))); err != nil {
		return fmt.Errorf("failed to apply role: %w", err)
	}

	log.Printf("Role %q reset successfully.", c.role)
	return nil
}

func (c *roleResetCmd) Usage() {
	executeUsage(c.fs.Output(), "role_reset_usage.txt", c)
}

func (c *roleResetCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleResetCmd)(nil)
