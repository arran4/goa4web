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
	args   []string
}

func parseBlogCommentsListCmd(parent *blogCommentsCmd, args []string) (*blogCommentsListCmd, error) {
	c := &blogCommentsListCmd{blogCommentsCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "blog id")
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
	uid := int32(c.UserID)
	b, err := queries.GetBlogEntryForUserById(ctx, dbpkg.GetBlogEntryForUserByIdParams{
		ViewerIdusers: uid,
		ID:            int32(c.ID),
	})
	if err != nil {
		return fmt.Errorf("get blog: %w", err)
	}
	var threadID int32
	if b.ForumthreadID.Valid {
		threadID = b.ForumthreadID.Int32
	}
	rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{
		UsersIdusers:   uid,
		UsersIdusers_2: uid,
		ForumthreadID:  threadID,
		UserID:         sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		return fmt.Errorf("list comments: %w", err)
	}
	for _, cm := range rows {
		fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	}
	return nil
}
