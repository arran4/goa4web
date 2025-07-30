package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// linksRefreshCmd implements "links refresh".
type linksRefreshCmd struct {
	*linksCmd
	fs *flag.FlagSet
	ID int
}

func parseLinksRefreshCmd(parent *linksCmd, args []string) (*linksRefreshCmd, error) {
	c := &linksRefreshCmd{linksCmd: parent}
	fs, _, err := parseFlags("refresh", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "link id")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *linksRefreshCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if err := queries.ClearExternalLinkCache(ctx, dbpkg.ClearExternalLinkCacheParams{UpdatedBy: sql.NullInt32{}, ID: int32(c.ID)}); err != nil {
		return fmt.Errorf("refresh link: %w", err)
	}
	return nil
}

func (c *linksRefreshCmd) Usage() {
	executeUsage(c.fs.Output(), "links_refresh_usage.txt", c)
}

func (c *linksRefreshCmd) FlagGroups() []flagGroup {
	return append(c.linksCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksRefreshCmd)(nil)
