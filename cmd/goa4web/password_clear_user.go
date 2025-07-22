package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// passwordClearUserCmd implements "password clear-user".
type passwordClearUserCmd struct {
	*passwordCmd
	fs       *flag.FlagSet
	Username string
	args     []string
}

func parsePasswordClearUserCmd(parent *passwordCmd, args []string) (*passwordClearUserCmd, error) {
	c := &passwordClearUserCmd{passwordCmd: parent}
	fs := flag.NewFlagSet("clear-user", flag.ContinueOnError)
	fs.StringVar(&c.Username, "username", "", "username")
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *passwordClearUserCmd) Run() error {
	if c.Username == "" {
		return fmt.Errorf("username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	user, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if err := queries.DeletePasswordResetsByUser(ctx, user.Idusers); err != nil {
		return fmt.Errorf("delete resets: %w", err)
	}
	return nil
}
