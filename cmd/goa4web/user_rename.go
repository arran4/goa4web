package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// userRenameCmd implements "user rename".
type userRenameCmd struct {
	*userCmd
	fs          *flag.FlagSet
	OldUsername string
	NewUsername string
}

func parseUserRenameCmd(parent *userCmd, args []string) (*userRenameCmd, error) {
	c := &userRenameCmd{userCmd: parent}
	fs, remaining, err := parseFlags("rename", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.OldUsername, "from", "", "The existing username to rename.")
		fs.StringVar(&c.NewUsername, "to", "", "The new username to assign.")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	if c.OldUsername == "" && len(remaining) > 0 {
		c.OldUsername = remaining[0]
		remaining = remaining[1:]
	}
	if c.NewUsername == "" && len(remaining) > 0 {
		c.NewUsername = remaining[0]
		remaining = remaining[1:]
	}
	if len(remaining) > 0 {
		fs.Usage()
		return nil, fmt.Errorf("too many arguments")
	}
	return c, nil
}

func (c *userRenameCmd) Run() error {
	if c.OldUsername == "" || c.NewUsername == "" {
		c.fs.Usage()
		return fmt.Errorf("from and to usernames required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	user, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.OldUsername, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if err := queries.AdminUpdateUsernameByID(ctx, db.AdminUpdateUsernameByIDParams{
		Username: sql.NullString{String: c.NewUsername, Valid: true},
		Idusers:  user.Idusers,
	}); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("username already exists")
		}
		return fmt.Errorf("rename user: %w", err)
	}
	c.rootCmd.Infof("renamed %s to %s", c.OldUsername, c.NewUsername)
	return nil
}
