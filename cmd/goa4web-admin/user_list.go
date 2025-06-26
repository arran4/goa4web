package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userListCmd implements "user list".
type userListCmd struct {
	*userCmd
	fs   *flag.FlagSet
	args []string
}

func parseUserListCmd(parent *userCmd, args []string) (*userListCmd, error) {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &userListCmd{userCmd: parent, fs: fs, args: fs.Args()}, nil
}

func (c *userListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	users, err := queries.AllUsers(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	for _, u := range users {
		fmt.Printf("%d\t%s\t%s\n", u.Idusers, u.Username.String, u.Email.String)
	}
	return nil
}
