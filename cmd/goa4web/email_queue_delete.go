package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// emailQueueDeleteCmd implements "email queue delete".
type emailQueueDeleteCmd struct {
	*emailQueueCmd
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseEmailQueueDeleteCmd(parent *emailQueueCmd, args []string) (*emailQueueDeleteCmd, error) {
	c := &emailQueueDeleteCmd{emailQueueCmd: parent}
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "pending email id")
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *emailQueueDeleteCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if err := queries.DeletePendingEmail(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("delete email: %w", err)
	}
	return nil
}
