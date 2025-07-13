package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// writingCommentsReadCmd implements "writing comments read".
type writingCommentsReadCmd struct {
	*writingCommentsCmd
	fs        *flag.FlagSet
	WritingID int
	CommentID int
	All       bool
	args      []string
}

func parseWritingCommentsReadCmd(parent *writingCommentsCmd, args []string) (*writingCommentsReadCmd, error) {
	c := &writingCommentsReadCmd{writingCommentsCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.WritingID, "id", 0, "writing id")
	fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	if c.WritingID == 0 && len(c.args) > 0 {
		if id, err := strconv.Atoi(c.args[0]); err == nil {
			c.WritingID = id
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

func (c *writingCommentsReadCmd) Run() error {
	if c.WritingID == 0 {
		return fmt.Errorf("writing id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	w, err := queries.GetWritingByIdForUserDescendingByPublishedDate(ctx, dbpkg.GetWritingByIdForUserDescendingByPublishedDateParams{Userid: 0, Idwriting: int32(c.WritingID)})
	if err != nil {
		return fmt.Errorf("get writing: %w", err)
	}
	if c.All {
		rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{
			UsersIdusers:   0,
			UsersIdusers_2: 0,
			ForumthreadID:  w.ForumthreadID,
			UserID:         sql.NullInt32{},
		})
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
	cm, err := queries.GetCommentByIdForUser(ctx, dbpkg.GetCommentByIdForUserParams{
		UsersIdusers:   0,
		UsersIdusers_2: 0,
		Idcomments:     int32(c.CommentID),
		UserID:         sql.NullInt32{},
	})
	if err != nil {
		return fmt.Errorf("get comment: %w", err)
	}
	fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	return nil
}
