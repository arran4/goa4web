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
	args      []string
}

func parseNewsCommentsReadCmd(parent *newsCommentsCmd, args []string) (*newsCommentsReadCmd, error) {
	c := &newsCommentsReadCmd{newsCommentsCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.NewsID, "id", 0, "news id")
	fs.IntVar(&c.CommentID, "comment", 0, "comment id")
	fs.IntVar(&c.UserID, "user", 0, "viewer user id")
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
	uid := int32(c.UserID)
	n, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCountForUser(ctx, dbpkg.GetNewsPostByIdWithWriterIdAndThreadCommentCountForUserParams{
		ViewerID: uid,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		ID:       int32(c.NewsID),
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
