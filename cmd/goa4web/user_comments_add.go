package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strings"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userCommentsAddCmd implements "user comments add".
type userCommentsAddCmd struct {
	*userCommentsCmd
	fs       *flag.FlagSet
	ID       int
	Username string
	Comment  string
}

func parseUserCommentsAddCmd(parent *userCommentsCmd, args []string) (*userCommentsAddCmd, error) {
	c := &userCommentsAddCmd{userCommentsCmd: parent}
	fs, _, err := parseFlags("add", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Comment, "comment", "", "comment text")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userCommentsAddCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	if strings.TrimSpace(c.Comment) == "" {
		return fmt.Errorf("empty comment")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if c.ID == 0 {
		u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	c.rootCmd.Verbosef("adding comment for user %d", c.ID)
	if err := queries.AdminInsertUserComment(ctx, dbpkg.AdminInsertUserCommentParams{UsersIdusers: int32(c.ID), Comment: c.Comment}); err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}
	c.rootCmd.Infof("added comment for user %d", c.ID)
	return nil
}
