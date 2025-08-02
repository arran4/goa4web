package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

// newsCommentsListCmd implements "news comments list".
type newsCommentsListCmd struct {
	*newsCommentsCmd
	fs     *flag.FlagSet
	ID     int
	UserID int
}

func parseNewsCommentsListCmd(parent *newsCommentsCmd, args []string) (*newsCommentsListCmd, error) {
	c := &newsCommentsListCmd{newsCommentsCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.IntVar(&c.ID, "id", 0, "news id")
	c.fs.IntVar(&c.UserID, "user", 0, "viewer user id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	rest := c.fs.Args()
	if c.ID == 0 && len(rest) > 0 {
		if id, err := strconv.Atoi(rest[0]); err == nil {
			c.ID = id
			rest = rest[1:]
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
	queries := db.New(db)
	uid := int32(c.UserID)
	n, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		ID:       int32(c.ID),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		return fmt.Errorf("get news: %w", err)
	}
	rows, err := queries.GetCommentsByThreadIdForUser(ctx, db.GetCommentsByThreadIdForUserParams{
		ViewerID: uid,
		ThreadID: n.ForumthreadID,
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
