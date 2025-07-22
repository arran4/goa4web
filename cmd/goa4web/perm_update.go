package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
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

func (c *permUpdateCmd) Run() error {
	if c.ID == 0 || c.Role == "" {
		return fmt.Errorf("id and role required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if err := queries.UpdatePermission(ctx, dbpkg.UpdatePermissionParams{
		IduserRoles: int32(c.ID),
		Name:        c.Role,
	}); err != nil {
		return fmt.Errorf("update permission: %w", err)
	}
	return nil
}
