package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

const defaultAuditLimit = 20 // number of rows shown when --limit not provided

// auditCmd implements the "audit" command.
type auditCmd struct {
	*rootCmd
	fs    *flag.FlagSet
	Limit int
	args  []string
}

func parseAuditCmd(parent *rootCmd, args []string) (*auditCmd, error) {
	c := &auditCmd{rootCmd: parent, Limit: defaultAuditLimit}
	fs := flag.NewFlagSet("audit", flag.ContinueOnError)
	fs.IntVar(&c.Limit, "limit", defaultAuditLimit, "number of rows to display")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *auditCmd) Run() error {
	db, err := c.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.GetRecentAuditLogs(ctx, int32(c.Limit))
	if err != nil {
		return fmt.Errorf("audit logs: %w", err)
	}
	for _, r := range rows {
		fmt.Printf("%d\t%s\t%s\t%s\n", r.ID, r.Username.String, r.Action, r.CreatedAt.Format(time.RFC3339))
	}
	return nil
}
