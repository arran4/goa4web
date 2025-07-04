package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// emailQueueListCmd implements "email queue list".
type emailQueueListCmd struct {
	*emailQueueCmd
	fs   *flag.FlagSet
	args []string
}

func parseEmailQueueListCmd(parent *emailQueueCmd, args []string) (*emailQueueListCmd, error) {
	c := &emailQueueListCmd{emailQueueCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *emailQueueListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.ListUnsentPendingEmails(ctx)
	if err != nil {
		return fmt.Errorf("list emails: %w", err)
	}
	for _, e := range rows {
		fmt.Printf("%d\t%s\t%s\t%s\n", e.ID, e.ToEmail, e.Subject, e.CreatedAt.Format(time.RFC3339))
	}
	return nil
}
