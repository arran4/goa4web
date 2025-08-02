package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// newsListCmd implements "news list".
type newsListCmd struct {
	*newsCmd
	fs     *flag.FlagSet
	Limit  int
	Offset int
}

func parseNewsListCmd(parent *newsCmd, args []string) (*newsListCmd, error) {
	c := &newsListCmd{newsCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.IntVar(&c.Limit, "limit", 10, "limit")
	c.fs.IntVar(&c.Offset, "offset", 0, "offset")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *newsListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(db)
	cd := common.NewCoreData(ctx, queries, c.rootCmd.cfg)
	posts, err := cd.LatestNewsList(int32(c.Offset), int32(c.Limit))
	if err != nil {
		return fmt.Errorf("list news: %w", err)
	}
	for _, n := range posts {
		fmt.Printf("%d\t%s\n", n.Idsitenews, n.News.String)
	}
	return nil
}
