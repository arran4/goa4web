package main

import (
	"context"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
)

// blogActivateCmd implements "blog activate".
type blogActivateCmd struct {
	*blogCmd
	fs *flag.FlagSet
	ID int
}

func parseBlogActivateCmd(parent *blogCmd, args []string) (*blogActivateCmd, error) {
	c := &blogActivateCmd{blogCmd: parent}
	c.fs = newFlagSet("activate")
	c.fs.IntVar(&c.ID, "id", 0, "blog id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *blogActivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	deactivated, err := queries.AdminIsBlogDeactivated(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if !deactivated {
		return fmt.Errorf("blog not deactivated")
	}
	rows, err := queries.AdminListDeactivatedBlogs(ctx, db.AdminListDeactivatedBlogsParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return fmt.Errorf("list deactivated: %w", err)
	}
	var found *db.AdminListDeactivatedBlogsRow
	for _, r := range rows {
		if int(r.Idblogs) == c.ID {
			found = r
			break
		}
	}
	if found == nil {
		return fmt.Errorf("blog %d not found", c.ID)
	}
	if err := queries.AdminRestoreBlog(ctx, db.AdminRestoreBlogParams{Blog: found.Blog, Idblogs: found.Idblogs}); err != nil {
		return fmt.Errorf("restore blog: %w", err)
	}
	if err := queries.AdminMarkBlogRestored(ctx, found.Idblogs); err != nil {
		return fmt.Errorf("mark restored: %w", err)
	}
	return nil
}
