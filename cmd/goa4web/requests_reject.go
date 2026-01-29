package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// requestsRejectCmd implements "requests reject".
type requestsRejectCmd struct {
	*requestsCmd
	fs      *flag.FlagSet
	request int
	comment string
	json    bool
}

func parseRequestsRejectCmd(parent *requestsCmd, args []string) (*requestsRejectCmd, error) {
	c := &requestsRejectCmd{requestsCmd: parent}
	fs, _, err := parseFlags("reject", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.request, "id", 0, "request ID")
		fs.StringVar(&c.comment, "comment", "", "optional comment")
		fs.BoolVar(&c.json, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *requestsRejectCmd) Usage() {
	executeUsage(c.fs.Output(), "requests_reject_usage.txt", c)
}

func (c *requestsRejectCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*requestsRejectCmd)(nil)

func (c *requestsRejectCmd) Run() error {
	if c.request == 0 {
		return fmt.Errorf("request id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	result, err := updateRequestStatus(ctx, queries, int32(c.request), "rejected", c.comment)
	if err != nil {
		return err
	}
	if c.json {
		return emitRequestActionJSON(result)
	}
	fmt.Printf("request %d rejected\n", c.request)
	return nil
}
