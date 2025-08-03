package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userCommentsListCmd implements "user comments list".
type userCommentsListCmd struct {
	*userCommentsCmd
	fs       *flag.FlagSet
	ID       int
	Username string
}

func parseUserCommentsListCmd(parent *userCommentsCmd, args []string) (*userCommentsListCmd, error) {
	c := &userCommentsListCmd{userCommentsCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userCommentsListCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(db)
	if c.ID == 0 {
		u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	rows, err := queries.ListAdminUserComments(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("list comments: %w", err)
	}
	for _, cm := range rows {
		fmt.Printf("%s\t%s\n", cm.CreatedAt.Format("2006-01-02 15:04"), cm.Comment)
	}
	return nil
}
