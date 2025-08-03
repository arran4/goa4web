package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// emailQueueDeleteCmd implements "email queue delete".
type emailQueueDeleteCmd struct {
	*emailQueueCmd
	fs *flag.FlagSet
	ID int
}

func parseEmailQueueDeleteCmd(parent *emailQueueCmd, args []string) (*emailQueueDeleteCmd, error) {
	c := &emailQueueDeleteCmd{emailQueueCmd: parent}
	c.fs = newFlagSet("delete")
	c.fs.IntVar(&c.ID, "id", 0, "pending email id")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailQueueDeleteCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if err := queries.AdminDeletePendingEmail(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("delete email: %w", err)
	}
	return nil
}
