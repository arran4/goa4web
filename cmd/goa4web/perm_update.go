package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// permUpdateCmd implements "perm update".
type permUpdateCmd struct {
	*permCmd
	fs   *flag.FlagSet
	ID   int
	Role string
}

func parsePermUpdateCmd(parent *permCmd, args []string) (*permUpdateCmd, error) {
	c := &permUpdateCmd{permCmd: parent}
	c.fs = newFlagSet("update")
	c.fs.IntVar(&c.ID, "id", 0, "permission id")
	c.fs.StringVar(&c.Role, "role", "", "permission role")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *permUpdateCmd) Usage() {
	executeUsage(c.fs.Output(), "perm_update_usage.txt", c)
}

func (c *permUpdateCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*permUpdateCmd)(nil)

func (c *permUpdateCmd) Run() error {
	if c.ID == 0 || c.Role == "" {
		return fmt.Errorf("id and role required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if err := queries.AdminUpdateUserRole(ctx, db.AdminUpdateUserRoleParams{
		IduserRoles: int32(c.ID),
		Name:        c.Role,
	}); err != nil {
		return fmt.Errorf("update permission: %w", err)
	}
	return nil
}
