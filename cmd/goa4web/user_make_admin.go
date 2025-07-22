package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userMakeAdminCmd grants administrator rights to an existing user.
type userMakeAdminCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
}

func parseUserMakeAdminCmd(parent *userCmd, args []string) (*userMakeAdminCmd, error) {
	c := &userMakeAdminCmd{userCmd: parent}
	fs, _, err := parseFlags("make-admin", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
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
	c.rootCmd.Verbosef("granting administrator to %s", c.Username)
	u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if _, err := queries.GetAdministratorUserRole(ctx, u.Idusers); err == nil {
		c.rootCmd.Verbosef("%s already administrator", c.Username)
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check admin: %w", err)
	}
	if err := queries.CreateUserRole(ctx, dbpkg.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         "administrator",
	}); err != nil {
		return fmt.Errorf("grant admin: %w", err)
	}
	c.rootCmd.Infof("granted administrator to %s", c.Username)
	return nil
}
