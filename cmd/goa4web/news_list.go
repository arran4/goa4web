package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// newsListCmd implements "news list".
type newsListCmd struct {
	*newsCmd
	fs     *flag.FlagSet
	Limit  int
	Offset int
	args   []string
}

func parseNewsListCmd(parent *newsCmd, args []string) (*newsListCmd, error) {
	c := &newsListCmd{newsCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.IntVar(&c.Limit, "limit", 10, "limit")
	fs.IntVar(&c.Offset, "offset", 0, "offset")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *newsListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(ctx, dbpkg.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
		ViewerID: 0,
		UserID:   sql.NullInt32{},
		Limit:    int32(c.Limit),
		Offset:   int32(c.Offset),
	})
	if err != nil {
		return fmt.Errorf("list news: %w", err)
	}
	for _, n := range rows {
		fmt.Printf("%d\t%s\n", n.Idsitenews, n.News.String)
	}
	return nil
}
