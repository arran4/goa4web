package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// linksDeleteCmd implements "links delete".
type linksDeleteCmd struct {
	*linksCmd
	fs *flag.FlagSet
	ID int
}

func parseLinksDeleteCmd(parent *linksCmd, args []string) (*linksDeleteCmd, error) {
	c := &linksDeleteCmd{linksCmd: parent}
	fs, _, err := parseFlags("delete", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "link id")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *linksDeleteCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if err := queries.AdminDeleteExternalLink(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	return nil
}

func (c *linksDeleteCmd) Usage() {
	executeUsage(c.fs.Output(), "links_delete_usage.txt", c)
}

func (c *linksDeleteCmd) FlagGroups() []flagGroup {
	return append(c.linksCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksDeleteCmd)(nil)
