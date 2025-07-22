package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// grantDeleteCmd implements "grant delete".
type grantDeleteCmd struct {
	*grantCmd
	fs *flag.FlagSet
	ID int
}

func parseGrantDeleteCmd(parent *grantCmd, args []string) (*grantDeleteCmd, error) {
	c := &grantDeleteCmd{grantCmd: parent}
	c.fs = newFlagSet("delete")
	c.fs.IntVar(&c.ID, "id", 0, "grant id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *grantDeleteCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	q := dbpkg.New(db)
	if err := q.DeleteGrant(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("delete grant: %w", err)
	}
	return nil
}
