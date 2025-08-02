package main

import (
	"context"
	"database/sql"
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
	UserID    int
	All       bool
}

func parseBlogCommentsReadCmd(parent *blogCommentsCmd, args []string) (*blogCommentsReadCmd, error) {
	c := &blogCommentsReadCmd{blogCommentsCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.BlogID, "id", 0, "blog id")
	c.fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	c.fs.IntVar(&c.UserID, "user", 0, "viewer user id")
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
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	uid := int32(c.UserID)
	b, err := queries.GetBlogEntryForListerByID(ctx, dbpkg.GetBlogEntryForListerByIDParams{
		ListerID: uid,
		ID:       int32(c.BlogID),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		return fmt.Errorf("get blog: %w", err)
	}
	if c.All {
		var threadID int32
		if b.ForumthreadID.Valid {
			threadID = b.ForumthreadID.Int32
		}
		rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{
			ViewerID: uid,
			ThreadID: threadID,
			UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
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
		ViewerID: uid,
		ID:       int32(c.CommentID),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		return fmt.Errorf("get comment: %w", err)
	}
	fmt.Printf("%d\t%s\n", cm.Idcomments, cm.Text.String)
	return nil
}
