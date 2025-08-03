package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

// newsCommentsReadCmd implements "news comments read".
type newsCommentsReadCmd struct {
	*newsCommentsCmd
	fs        *flag.FlagSet
	NewsID    int
	CommentID int
	All       bool
}

func parseNewsCommentsReadCmd(parent *newsCommentsCmd, args []string) (*newsCommentsReadCmd, error) {
	c := &newsCommentsReadCmd{newsCommentsCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.NewsID, "id", 0, "news id")
	c.fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	rest := c.fs.Args()
	if c.NewsID == 0 && len(rest) > 0 {
		if id, err := strconv.Atoi(rest[0]); err == nil {
			c.NewsID = id
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

func (c *newsCommentsReadCmd) Run() error {
	if c.NewsID == 0 {
		return fmt.Errorf("news id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	n, err := queries.SystemGetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, int32(c.NewsID))
	if err != nil {
		return fmt.Errorf("get news: %w", err)
	}
	if c.All {
		rows, err := queries.SystemListCommentsByThreadID(ctx, n.ForumthreadID)
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
