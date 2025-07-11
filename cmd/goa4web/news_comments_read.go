package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// newsCommentsReadCmd implements "news comments read".
type newsCommentsReadCmd struct {
	*newsCommentsCmd
	fs        *flag.FlagSet
	NewsID    int
	CommentID int
	All       bool
	args      []string
}

func parseNewsCommentsReadCmd(parent *newsCommentsCmd, args []string) (*newsCommentsReadCmd, error) {
	c := &newsCommentsReadCmd{newsCommentsCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.NewsID, "id", 0, "news id")
	fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	if c.NewsID == 0 && len(c.args) > 0 {
		if id, err := strconv.Atoi(c.args[0]); err == nil {
			c.NewsID = id
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

func (c *newsCommentsReadCmd) Run() error {
	if c.NewsID == 0 {
		return fmt.Errorf("news id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	n, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, int32(c.NewsID))
	if err != nil {
		return fmt.Errorf("get news: %w", err)
	}
	if c.All {
		rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{UsersIdusers: 0, ForumthreadID: n.ForumthreadID})
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
