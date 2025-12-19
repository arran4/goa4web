package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/db"
)

// roleRemoveCmd implements the "role remove" subcommand.
type roleRemoveCmd struct {
	*roleCmd
	fs       *flag.FlagSet
	roleName string
}

func parseRoleRemoveCmd(parent *roleCmd, args []string) (*roleRemoveCmd, error) {
	c := &roleRemoveCmd{roleCmd: parent}
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.roleName, "name", "", "The role name to remove.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.roleName == "" {
		if fs.NArg() > 0 {
			c.roleName = fs.Arg(0)
		} else {
			return nil, fmt.Errorf("role name is required")
		}
	}
	return c, nil
}

func (c *roleRemoveCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	ctx := c.rootCmd.ctx

	// Start transaction
	tx, err := sdb.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	q := db.New(tx)

	role, err := q.GetRoleByName(ctx, c.roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("role %q not found", c.roleName)
		}
		return fmt.Errorf("failed to get role by name: %w", err)
	}

	// Delete grants first
	if err := q.DeleteGrantsByRoleID(ctx, sql.NullInt32{Int32: role.ID, Valid: true}); err != nil {
		return fmt.Errorf("failed to delete grants for role: %w", err)
	}

	// Delete role
	// Assuming raw SQL as delete role query might not be in sqlc
	_, err = tx.ExecContext(ctx, "DELETE FROM roles WHERE id = ?", role.ID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	log.Printf("Role %q removed.", c.roleName)
	return nil
}

func (c *roleRemoveCmd) Usage() {
	executeUsage(c.fs.Output(), "role_remove_usage.txt", c)
}

func (c *roleRemoveCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleRemoveCmd)(nil)
