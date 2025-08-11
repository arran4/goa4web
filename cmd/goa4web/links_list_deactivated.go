package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// linksListDeactivatedCmd implements "links list-deactivated".
type linksListDeactivatedCmd struct {
	*linksCmd
	fs     *flag.FlagSet
	Limit  int
	Offset int
}

func parseLinksListDeactivatedCmd(parent *linksCmd, args []string) (*linksListDeactivatedCmd, error) {
	c := &linksListDeactivatedCmd{linksCmd: parent}
	c.fs = newFlagSet("list-deactivated")
	c.fs.IntVar(&c.Limit, "limit", 20, "max results")
	c.fs.IntVar(&c.Offset, "offset", 0, "result offset")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *linksListDeactivatedCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.AdminListDeactivatedLinks(ctx, db.AdminListDeactivatedLinksParams{Limit: int32(c.Limit), Offset: int32(c.Offset)})
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	for _, r := range rows {
		fmt.Printf("%d\t%s\t%s\n", r.ID, r.Title.String, r.Url.String)
	}
	return nil
}
