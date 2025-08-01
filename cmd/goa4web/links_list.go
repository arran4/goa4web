package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// linksListCmd implements "links list".
type linksListCmd struct {
	*linksCmd
	fs *flag.FlagSet
}

func parseLinksListCmd(parent *linksCmd, args []string) (*linksListCmd, error) {
	c := &linksListCmd{linksCmd: parent}
	fs, _, err := parseFlags("list", args, nil)
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *linksListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.ListExternalLinksForAdmin(ctx, dbpkg.ListExternalLinksForAdminParams{Limit: 200, Offset: 0})
	if err != nil {
		return fmt.Errorf("list links: %w", err)
	}
	for _, l := range rows {
		fmt.Printf("%d\t%s\t%d\t%s\n", l.ID, l.Url, l.Clicks, l.CreatedAt)
	}
	return nil
}

func (c *linksListCmd) Usage() {
	executeUsage(c.fs.Output(), "links_list_usage.txt", c)
}

func (c *linksListCmd) FlagGroups() []flagGroup {
	return append(c.linksCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksListCmd)(nil)
