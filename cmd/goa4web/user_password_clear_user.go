package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userPasswordClearUserCmd implements "user password clear-user".
// It deletes reset password requests for a specific user.
type userPasswordClearUserCmd struct {
	*userPasswordCmd
	fs       *flag.FlagSet
	Username string
	args     []string
}

func parseUserPasswordClearUserCmd(parent *userPasswordCmd, args []string) (*userPasswordClearUserCmd, error) {
	c := &userPasswordClearUserCmd{userPasswordCmd: parent}
	fs := flag.NewFlagSet("clear-user", flag.ContinueOnError)
	fs.StringVar(&c.Username, "username", "", "username")
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *userPasswordClearUserCmd) Run() error {
	if c.Username == "" {
		return fmt.Errorf("username required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	user, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	res, err := queries.DeletePasswordResetsByUser(ctx, user.Idusers)
	if err != nil {
		return fmt.Errorf("delete resets: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil {
		c.rootCmd.Infof("deleted %d password reset requests", rows)
	}
	return nil
}
