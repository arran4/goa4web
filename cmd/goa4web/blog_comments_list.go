package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// blogCommentsListCmd implements "blog comments list".
type blogCommentsListCmd struct {
	*blogCommentsCmd
	fs     *flag.FlagSet
	ID     int
	UserID int
}

func parseBlogCommentsListCmd(parent *blogCommentsCmd, args []string) (*blogCommentsListCmd, error) {
	c := &blogCommentsListCmd{blogCommentsCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.IntVar(&c.ID, "id", 0, "blog id")
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

func (c *blogCommentsListCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	listerID := int32(c.UserID)
	b, err := queries.GetBlogEntryForListerByID(ctx, dbpkg.GetBlogEntryForListerByIDParams{
		ListerID: listerID,
		ID:       int32(c.ID),
		UserID:   sql.NullInt32{Int32: listerID, Valid: listerID != 0},
	})
	if err != nil {
		return fmt.Errorf("get blog: %w", err)
	}
	var threadID int32
	if b.ForumthreadID.Valid {
		threadID = b.ForumthreadID.Int32
	}
	rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{
		ViewerID: listerID,
		ThreadID: threadID,
		UserID:   sql.NullInt32{Int32: listerID, Valid: listerID != 0},
	})
	if err != nil {
		return fmt.Errorf("list comments: %w", err)
	}
	for _, cm := range rows {
		fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	}
	return nil
}
