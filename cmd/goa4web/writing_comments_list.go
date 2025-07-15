package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// writingCommentsListCmd implements "writing comments list".
type writingCommentsListCmd struct {
	*writingCommentsCmd
	fs     *flag.FlagSet
	ID     int
	UserID int
	args   []string
}

func parseWritingCommentsListCmd(parent *writingCommentsCmd, args []string) (*writingCommentsListCmd, error) {
	c := &writingCommentsListCmd{writingCommentsCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "writing id")
	fs.IntVar(&c.UserID, "user", 0, "viewer user id")
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

func (c *writingCommentsListCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	uid := int32(c.UserID)
	w, err := queries.GetWritingByIdForUserDescendingByPublishedDate(ctx, dbpkg.GetWritingByIdForUserDescendingByPublishedDateParams{ViewerID: uid, Idwriting: int32(c.ID), ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0}})
	if err != nil {
		return fmt.Errorf("get writing: %w", err)
	}
	rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: w.ForumthreadID,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		return fmt.Errorf("list comments: %w", err)
	}
	for _, cm := range rows {
		fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	}
	return nil
}
