package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/roles"
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

	data, err := roles.ReadRoleSQL(c.role, c.file)
	if err != nil {
		return err
	}
	if err := roles.ApplyRoleSQL(context.Background(), c.role, data, sdb); err != nil {
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
