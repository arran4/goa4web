package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// requestsCommentCmd implements "requests comment".
type requestsCommentCmd struct {
	*requestsCmd
	fs      *flag.FlagSet
	request int
	comment string
	json    bool
}

func parseRequestsCommentCmd(parent *requestsCmd, args []string) (*requestsCommentCmd, error) {
	c := &requestsCommentCmd{requestsCmd: parent}
	fs, _, err := parseFlags("comment", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.request, "id", 0, "request ID")
		fs.StringVar(&c.comment, "comment", "", "comment text")
		fs.BoolVar(&c.json, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *requestsCommentCmd) Usage() {
	executeUsage(c.fs.Output(), "requests_comment_usage.txt", c)
}

func (c *requestsCommentCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*requestsCommentCmd)(nil)

func (c *requestsCommentCmd) Run() error {
	if c.request == 0 {
		return fmt.Errorf("request id required")
	}
	if c.comment == "" {
		return fmt.Errorf("comment required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if _, err := queries.AdminGetRequestByID(ctx, int32(c.request)); err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	if err := queries.AdminInsertRequestComment(ctx, db.AdminInsertRequestCommentParams{
		RequestID: int32(c.request),
		Comment:   c.comment,
	}); err != nil {
		return fmt.Errorf("insert request comment: %w", err)
	}
	if c.json {
		b, _ := json.MarshalIndent(requestActionResult{
			ID:      int32(c.request),
			Comment: c.comment,
		}, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	fmt.Printf("comment added to request %d\n", c.request)
	return nil
}
