package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// writingListDeactivatedCmd implements "writing list-deactivated".
type writingListDeactivatedCmd struct {
	*writingCmd
	fs     *flag.FlagSet
	Limit  int
	Offset int
}

func parseWritingListDeactivatedCmd(parent *writingCmd, args []string) (*writingListDeactivatedCmd, error) {
	c := &writingListDeactivatedCmd{writingCmd: parent}
	c.fs = newFlagSet("list-deactivated")
	c.fs.IntVar(&c.Limit, "limit", 20, "max results")
	c.fs.IntVar(&c.Offset, "offset", 0, "result offset")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *writingListDeactivatedCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.AdminListDeactivatedWritings(ctx, db.AdminListDeactivatedWritingsParams{Limit: int32(c.Limit), Offset: int32(c.Offset)})
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	for _, r := range rows {
		fmt.Printf("%d\t%s\n", r.Idwriting, r.Title.String)
	}
	return nil
}
