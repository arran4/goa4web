package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// permRevokeCmd implements "perm revoke".
type permRevokeCmd struct {
	*permCmd
	fs *flag.FlagSet
	ID int
}

func parsePermRevokeCmd(parent *permCmd, args []string) (*permRevokeCmd, error) {
	c := &permRevokeCmd{permCmd: parent}
	c.fs = newFlagSet("revoke")
	c.fs.IntVar(&c.ID, "id", 0, "The ID of the permission to revoke. You can get the ID by running 'perm list'.")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *permRevokeCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if err := queries.AdminDeleteUserRole(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("revoke: %w", err)
	}
	return nil
}
