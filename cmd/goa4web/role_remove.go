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
	srcRole  string
	destRole string
}

func parseRoleRemoveCmd(parent *roleCmd, args []string) (*roleRemoveCmd, error) {
	c := &roleRemoveCmd{roleCmd: parent}
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.srcRole, "src", "", "The source role.")
	fs.StringVar(&c.destRole, "dest", "", "The destination role.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.destRole == "" {
		return nil, fmt.Errorf("destination role is required")
	}
	return c, nil
}

func (c *roleRemoveCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.New(sdb)
	ctx := c.rootCmd.ctx

	destRole, err := q.GetRoleByName(ctx, c.destRole)
	if err != nil {
		return fmt.Errorf("failed to get destination role by name: %w", err)
	}

	if c.srcRole == "" {
		log.Printf("Removing all grants from %q", c.destRole)
		if err := q.DeleteGrantsByRoleID(ctx, sql.NullInt32{Int32: destRole.ID, Valid: true}); err != nil {
			return fmt.Errorf("failed to delete grants: %w", err)
		}
		log.Printf("Successfully removed all grants from %q.", c.destRole)
		return nil
	}

	srcRole, err := q.GetRoleByName(ctx, c.srcRole)
	if err != nil {
		return fmt.Errorf("failed to get source role by name: %w", err)
	}

	grants, err := q.GetGrantsByRoleID(ctx, sql.NullInt32{Int32: srcRole.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get grants for source role: %w", err)
	}

	log.Printf("Removing %d grants from %q that match %q", len(grants), c.destRole, c.srcRole)
	for _, grant := range grants {
		params := db.DeleteGrantByPropertiesParams{
			RoleID:  sql.NullInt32{Int32: destRole.ID, Valid: true},
			Section: grant.Section,
			Item:    grant.Item,
			Action:  grant.Action,
		}
		if err := q.DeleteGrantByProperties(ctx, params); err != nil {
			return fmt.Errorf("failed to delete grant: %w", err)
		}
	}

	log.Printf("Successfully removed grants from %q that match %q.", c.destRole, c.srcRole)
	return nil
}

func (c *roleRemoveCmd) Usage() {
	executeUsage(c.fs.Output(), "role_remove_usage.txt", c)
}

func (c *roleRemoveCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleRemoveCmd)(nil)
