package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// blogReadCmd implements "blog read".
type blogReadCmd struct {
	*blogCmd
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseBlogReadCmd(parent *blogCmd, args []string) (*blogReadCmd, error) {
	c := &blogReadCmd{blogCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "blog id")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	if c.ID == 0 && len(c.args) > 0 {
		if id, err := strconv.Atoi(c.args[0]); err == nil {
			c.ID = id
			c.args = c.args[1:]
		}
	}
	return c, nil
}

func (c *blogReadCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	row, err := queries.GetBlogEntryForUserById(ctx, dbpkg.GetBlogEntryForUserByIdParams{
		ViewerIdusers: 0,
		ID:            int32(c.ID),
	})
	if err != nil {
		return fmt.Errorf("get blog: %w", err)
	}
	fmt.Printf("Written: %s\n", row.Written.Format(time.RFC3339))
	fmt.Println(row.Blog.String)
	return nil
}
