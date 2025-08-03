package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// blogListCmd implements "blog list".
type blogListCmd struct {
	*blogCmd
	fs     *flag.FlagSet
	UserID int
	Limit  int
	Offset int
}

func parseBlogListCmd(parent *blogCmd, args []string) (*blogListCmd, error) {
	c := &blogListCmd{blogCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.IntVar(&c.UserID, "user", 0, "user id")
	c.fs.IntVar(&c.Limit, "limit", 10, "limit")
	c.fs.IntVar(&c.Offset, "offset", 0, "offset")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *blogListCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	uid := int32(c.UserID)
	rows, err := queries.ListBlogEntriesForLister(ctx, db.ListBlogEntriesForListerParams{
		ListerID: uid,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		Limit:    int32(c.Limit),
		Offset:   int32(c.Offset),
	})
	if err != nil {
		return fmt.Errorf("list blogs: %w", err)
	}
	for _, b := range rows {
		fmt.Printf("%d\t%s\n", b.Idblogs, b.Blog.String)
	}
	return nil
}
