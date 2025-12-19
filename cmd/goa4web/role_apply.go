package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/db"
)

// roleApplyCmd implements the "role apply" subcommand.
type roleApplyCmd struct {
	*roleCmd
	fs       *flag.FlagSet
	srcRole  string
	destRole string
}

func parseRoleApplyCmd(parent *roleCmd, args []string) (*roleApplyCmd, error) {
	c := &roleApplyCmd{roleCmd: parent}
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.srcRole, "src", "", "The source role.")
	fs.StringVar(&c.destRole, "dest", "", "The destination role.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.srcRole == "" || c.destRole == "" {
		return nil, fmt.Errorf("source and destination roles are required")
	}
	return c, nil
}

func (c *roleApplyCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.New(sdb)
	ctx := c.rootCmd.ctx

	srcRole, err := q.GetRoleByName(ctx, c.srcRole)
	if err != nil {
		return fmt.Errorf("failed to get source role by name: %w", err)
	}

	destRole, err := q.GetRoleByName(ctx, c.destRole)
	if err != nil {
		return fmt.Errorf("failed to get destination role by name: %w", err)
	}

	grants, err := q.GetGrantsByRoleID(ctx, sql.NullInt32{Int32: srcRole.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get grants for source role: %w", err)
	}

	log.Printf("Applying %d grants from %q to %q", len(grants), c.srcRole, c.destRole)
	for _, grant := range grants {
		params := db.CreateGrantParams{
			RoleID:   sql.NullInt32{Int32: destRole.ID, Valid: true},
			Section:  grant.Section,
			Item:     grant.Item,
			RuleType: grant.RuleType,
			ItemID:   grant.ItemID,
			ItemRule: grant.ItemRule,
			Action:   grant.Action,
			Extra:    grant.Extra,
			Active:   grant.Active,
		}
		if err := q.CreateGrant(ctx, params); err != nil {
			return fmt.Errorf("failed to create grant: %w", err)
		}
	}

	log.Printf("Successfully applied grants from %q to %q.", c.srcRole, c.destRole)
	return nil
}

func (c *roleApplyCmd) Usage() {
	executeUsage(c.fs.Output(), "role_apply_usage.txt", c)
}

func (c *roleApplyCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleApplyCmd)(nil)
