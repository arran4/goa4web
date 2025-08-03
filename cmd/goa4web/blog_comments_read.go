package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

// blogCommentsReadCmd implements "blog comments read".
type blogCommentsReadCmd struct {
	*blogCommentsCmd
	fs        *flag.FlagSet
	BlogID    int
	CommentID int
	All       bool
}

func parseBlogCommentsReadCmd(parent *blogCommentsCmd, args []string) (*blogCommentsReadCmd, error) {
	c := &blogCommentsReadCmd{blogCommentsCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.BlogID, "id", 0, "blog id")
	c.fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	rest := c.fs.Args()
	if c.BlogID == 0 && len(rest) > 0 {
		if id, err := strconv.Atoi(rest[0]); err == nil {
			c.BlogID = id
			rest = rest[1:]
		}
	}
	if len(rest) > 0 {
		if rest[0] == "all" {
			c.All = true
			rest = rest[1:]
		} else if id, err := strconv.Atoi(rest[0]); err == nil {
			c.CommentID = id
			rest = rest[1:]
		}
	}
	return c, nil
}

func (c *blogCommentsReadCmd) Run() error {
	if c.BlogID == 0 {
		return fmt.Errorf("blog id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	b, err := queries.SystemGetBlogEntryByID(ctx, int32(c.BlogID))
	if err != nil {
		return fmt.Errorf("get blog: %w", err)
	}
	if c.All {
		var threadID int32
		if b.ForumthreadID.Valid {
			threadID = b.ForumthreadID.Int32
		}
		rows, err := queries.SystemListCommentsByThreadID(ctx, threadID)
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
	cm, err := queries.GetCommentById(ctx, int32(c.CommentID))
	if err != nil {
		return fmt.Errorf("get comment: %w", err)
	}
	fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	return nil
}
