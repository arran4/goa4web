package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// requestsAcceptCmd implements "requests accept".
type requestsAcceptCmd struct {
	*requestsCmd
	fs      *flag.FlagSet
	request int
	comment string
	json    bool
}

func parseRequestsAcceptCmd(parent *requestsCmd, args []string) (*requestsAcceptCmd, error) {
	c := &requestsAcceptCmd{requestsCmd: parent}
	fs, _, err := parseFlags("accept", args, func(fs *flag.FlagSet) {
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

func (c *requestsAcceptCmd) Usage() {
	executeUsage(c.fs.Output(), "requests_accept_usage.txt", c)
}

func (c *requestsAcceptCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*requestsAcceptCmd)(nil)

func (c *requestsAcceptCmd) Run() error {
	if c.request == 0 {
		return fmt.Errorf("request id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	result, err := updateRequestStatus(ctx, queries, int32(c.request), "accepted", c.comment)
	if err != nil {
		return err
	}
	if c.json {
		return emitRequestActionJSON(result)
	}
	fmt.Printf("request %d accepted\n", c.request)
	return nil
}
