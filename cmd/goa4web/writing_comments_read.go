package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

// writingCommentsReadCmd implements "writing comments read".
type writingCommentsReadCmd struct {
	*writingCommentsCmd
	fs        *flag.FlagSet
	WritingID int
	CommentID int
	All       bool
}

func parseWritingCommentsReadCmd(parent *writingCommentsCmd, args []string) (*writingCommentsReadCmd, error) {
	c := &writingCommentsReadCmd{writingCommentsCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.WritingID, "id", 0, "writing id")
	c.fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	rest := c.fs.Args()
	if c.WritingID == 0 && len(rest) > 0 {
		if id, err := strconv.Atoi(rest[0]); err == nil {
			c.WritingID = id
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

func (c *writingCommentsReadCmd) Run() error {
	if c.WritingID == 0 {
		return fmt.Errorf("writing id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	w, err := queries.SystemGetWritingByID(ctx, int32(c.WritingID))
	if err != nil {
		return fmt.Errorf("get writing: %w", err)
	}
	if c.All {
		rows, err := queries.SystemListCommentsByThreadID(ctx, w.ForumthreadID)
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
