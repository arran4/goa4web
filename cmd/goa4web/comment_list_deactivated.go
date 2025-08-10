package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// commentListDeactivatedCmd implements "comment list-deactivated".
type commentListDeactivatedCmd struct {
	*commentCmd
	fs     *flag.FlagSet
	Limit  int
	Offset int
}

func parseCommentListDeactivatedCmd(parent *commentCmd, args []string) (*commentListDeactivatedCmd, error) {
	c := &commentListDeactivatedCmd{commentCmd: parent}
	c.fs = newFlagSet("list-deactivated")
	c.fs.IntVar(&c.Limit, "limit", 20, "max results")
	c.fs.IntVar(&c.Offset, "offset", 0, "result offset")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *commentListDeactivatedCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.AdminListDeactivatedComments(ctx, db.AdminListDeactivatedCommentsParams{Limit: int32(c.Limit), Offset: int32(c.Offset)})
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	for _, r := range rows {
		fmt.Printf("%d\t%s\n", r.Idcomments, r.Text.String)
	}
	return nil
}
