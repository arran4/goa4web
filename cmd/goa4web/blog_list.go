package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// blogListCmd implements "blog list".
type blogListCmd struct {
	*blogCmd
	fs     *flag.FlagSet
	UserID int
	Limit  int
	Offset int
	args   []string
}

func parseBlogListCmd(parent *blogCmd, args []string) (*blogListCmd, error) {
	c := &blogListCmd{blogCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.IntVar(&c.UserID, "user", 0, "user id")
	fs.IntVar(&c.Limit, "limit", 10, "limit")
	fs.IntVar(&c.Offset, "offset", 0, "offset")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *blogListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.GetBlogEntriesForUserDescending(ctx, dbpkg.GetBlogEntriesForUserDescendingParams{
		UsersIdusers:       int32(c.UserID),
		LanguageIdlanguage: 0,
		Limit:              int32(c.Limit),
		Offset:             int32(c.Offset),
	})
	if err != nil {
		return fmt.Errorf("list blogs: %w", err)
	}
	for _, b := range rows {
		fmt.Printf("%d\t%s\n", b.Idblogs, b.Blog.String)
	}
	return nil
}
