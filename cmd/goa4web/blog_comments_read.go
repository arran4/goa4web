package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// blogCommentsReadCmd implements "blog comments read".
type blogCommentsReadCmd struct {
	*blogCommentsCmd
	fs        *flag.FlagSet
	BlogID    int
	CommentID int
	All       bool
	args      []string
}

func parseBlogCommentsReadCmd(parent *blogCommentsCmd, args []string) (*blogCommentsReadCmd, error) {
	c := &blogCommentsReadCmd{blogCommentsCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.BlogID, "id", 0, "blog id")
	fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	if c.BlogID == 0 && len(c.args) > 0 {
		if id, err := strconv.Atoi(c.args[0]); err == nil {
			c.BlogID = id
			c.args = c.args[1:]
		}
	}
	if len(c.args) > 0 {
		if c.args[0] == "all" {
			c.All = true
			c.args = c.args[1:]
		} else if id, err := strconv.Atoi(c.args[0]); err == nil {
			c.CommentID = id
			c.args = c.args[1:]
		}
	}
	return c, nil
}

func (c *blogCommentsReadCmd) Run() error {
	if c.BlogID == 0 {
		return fmt.Errorf("blog id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	b, err := queries.GetBlogEntryForUserById(ctx, int32(c.BlogID))
	if err != nil {
		return fmt.Errorf("get blog: %w", err)
	}
	if c.All {
		rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{UsersIdusers: 0, ForumthreadIdforumthread: b.ForumthreadIdforumthread})
		if err != nil {
			return fmt.Errorf("get comments: %w", err)
		}
		for _, cm := range rows {
			fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
		}
		return nil
	}
	if c.CommentID == 0 {
		return fmt.Errorf("comment id required")
	}
	cm, err := queries.GetCommentByIdForUser(ctx, dbpkg.GetCommentByIdForUserParams{UsersIdusers: 0, Idcomments: int32(c.CommentID)})
	if err != nil {
		return fmt.Errorf("get comment: %w", err)
	}
	fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	return nil
}
