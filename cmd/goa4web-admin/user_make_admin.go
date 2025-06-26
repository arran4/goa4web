package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userMakeAdminCmd grants administrator rights to an existing user.
type userMakeAdminCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	args     []string
}

func parseUserMakeAdminCmd(parent *userCmd, args []string) (*userMakeAdminCmd, error) {
	c := &userMakeAdminCmd{userCmd: parent}
	fs := flag.NewFlagSet("make-admin", flag.ContinueOnError)
	fs.StringVar(&c.Username, "username", "", "username")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *userMakeAdminCmd) Run() error {
	if c.Username == "" {
		return fmt.Errorf("username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if err := queries.PermissionUserAllow(ctx, dbpkg.PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section:      sql.NullString{String: "administrator", Valid: true},
		Level:        sql.NullString{String: "administrator", Valid: true},
	}); err != nil {
		return fmt.Errorf("grant admin: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("granted administrator to %s\n", c.Username)
	}
	return nil
}
