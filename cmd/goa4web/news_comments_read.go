package main

import (
	"context"
	"database/sql"
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
	UserID    int
	All       bool
}

func parseNewsCommentsReadCmd(parent *newsCommentsCmd, args []string) (*newsCommentsReadCmd, error) {
	c := &newsCommentsReadCmd{newsCommentsCmd: parent}
	c.fs = newFlagSet("read")
	c.fs.IntVar(&c.NewsID, "id", 0, "news id")
	c.fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	c.fs.IntVar(&c.UserID, "user", 0, "viewer user id")
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
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	uid := int32(c.UserID)
	n, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx, dbpkg.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: uid,
		ID:       int32(c.NewsID),
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		return fmt.Errorf("get news: %w", err)
	}
	if c.All {
		rows, err := queries.GetCommentsByThreadIdForUser(ctx, dbpkg.GetCommentsByThreadIdForUserParams{
			ViewerID: uid,
			ThreadID: n.ForumthreadID,
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
