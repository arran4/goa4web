package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// grantListCmd implements "grant list".
type grantListCmd struct {
	*grantCmd
	fs *flag.FlagSet
}

func parseGrantListCmd(parent *grantCmd, args []string) (*grantListCmd, error) {
	c := &grantListCmd{grantCmd: parent}
	c.fs = newFlagSet("list")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *grantListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	q := db.New(db)
	rows, err := q.ListGrants(ctx)
	if err != nil {
		return fmt.Errorf("list grants: %w", err)
	}
	for _, g := range rows {
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n", g.ID, g.Section, g.Item.String, g.Action, g.RuleType)
	}
	return nil
}
