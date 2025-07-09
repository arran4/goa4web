package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// newsCommentsListCmd implements "news comments list".
type newsCommentsListCmd struct {
	*newsCommentsCmd
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseNewsCommentsListCmd(parent *newsCommentsCmd, args []string) (*newsCommentsListCmd, error) {
	c := &newsCommentsListCmd{newsCommentsCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "news id")
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

func (c *newsCommentsListCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	n, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("get news: %w", err)
	}
	rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{UsersIdusers: 0, ForumthreadID: n.ForumthreadID})
	if err != nil {
		return fmt.Errorf("list comments: %w", err)
	}
	for _, cm := range rows {
		fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	}
	return nil
}
