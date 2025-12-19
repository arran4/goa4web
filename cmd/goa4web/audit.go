package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

const defaultAuditLimit = 20 // number of rows shown when --limit not provided

// auditCmd implements the "audit" command.
type auditCmd struct {
	*rootCmd
	fs    *flag.FlagSet
	Limit int
}

func parseAuditCmd(parent *rootCmd, args []string) (*auditCmd, error) {
	c := &auditCmd{rootCmd: parent, Limit: defaultAuditLimit}
	c.fs = newFlagSet("audit")
	c.fs.IntVar(&c.Limit, "limit", defaultAuditLimit, "number of rows to display")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *auditCmd) Run() error {
	sdb, err := c.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(sdb)
	rows, err := queries.AdminGetRecentAuditLogs(ctx, int32(c.Limit))
	if err != nil {
		return fmt.Errorf("audit logs: %w", err)
	}
	for _, r := range rows {
		fmt.Printf("%d\t%s\t%s\t%s\n", r.ID, r.Username.String, r.Action, r.CreatedAt.Format(time.RFC3339))
	}
	return nil
}

func (c *auditCmd) Usage() {
	executeUsage(c.fs.Output(), "audit_usage.txt", c)
}

func (c *auditCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*auditCmd)(nil)
