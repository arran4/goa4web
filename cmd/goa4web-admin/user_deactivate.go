package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userDeactivateCmd implements "user deactivate" which anonymises an account.
type userDeactivateCmd struct {
	*userCmd
	fs       *flag.FlagSet
	ID       int
	Username string
	args     []string
}

func parseUserDeactivateCmd(parent *userCmd, args []string) (*userDeactivateCmd, error) {
	c := &userDeactivateCmd{userCmd: parent}
	fs := flag.NewFlagSet("deactivate", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "user id")
	fs.StringVar(&c.Username, "username", "", "username")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *userDeactivateCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	var uid int32
	if c.ID != 0 {
		uid = int32(c.ID)
	} else {
		u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		uid = u.Idusers
	}
	if err := queries.DeactivateUser(ctx, uid); err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("deactivated user %d\n", uid)
	}
	return nil
}
