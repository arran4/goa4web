package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// blogReadCmd implements "blog read".
type blogReadCmd struct {
	*blogCmd
	fs *flag.FlagSet
	ID int
}

func parseBlogReadCmd(parent *blogCmd, args []string) (*blogReadCmd, error) {
	c := &blogReadCmd{blogCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.ID, "id", 0, "blog id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	rest := c.fs.Args()
	if c.ID == 0 && len(rest) > 0 {
		if id, err := strconv.Atoi(rest[0]); err == nil {
			c.ID = id
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
	row, err := queries.GetBlogEntryForViewerById(ctx, dbpkg.GetBlogEntryForViewerByIdParams{
		ViewerID: 0,
		UserID:   sql.NullInt32{Int32: 0, Valid: false},
		ID:       int32(c.ID),
	})
	if err != nil {
		return fmt.Errorf("get blog: %w", err)
	}
	fmt.Printf("Written: %s\n", row.Written.Format(time.RFC3339))
	fmt.Println(row.Blog.String)
	return nil
}
