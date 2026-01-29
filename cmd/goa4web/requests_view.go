package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

// requestsViewCmd implements "requests view".
type requestsViewCmd struct {
	*requestsCmd
	fs *flag.FlagSet
	ID int
}

type requestViewOutput struct {
	Request requestJSON `json:"request"`
}

func parseRequestsViewCmd(parent *requestsCmd, args []string) (*requestsViewCmd, error) {
	c := &requestsViewCmd{requestsCmd: parent}
	c.fs = newFlagSet("view")
	c.fs.IntVar(&c.ID, "id", 0, "request id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	rest := c.fs.Args()
	if c.ID == 0 && len(rest) > 0 {
		if id, err := strconv.Atoi(rest[0]); err == nil {
			c.ID = id
		}
	}
	return c, nil
}

func (c *requestsViewCmd) Usage() {
	executeUsage(c.fs.Output(), "requests_view_usage.txt", c)
}

func (c *requestsViewCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*requestsViewCmd)(nil)

func (c *requestsViewCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	row, err := queries.AdminGetRequestByID(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}

	payload := requestViewOutput{Request: requestToJSON(row)}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
