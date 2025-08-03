package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userAddRoleCmd implements the "user add-role" command.
type userAddRoleCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	Role     string
}

func parseUserAddRoleCmd(parent *userCmd, args []string) (*userAddRoleCmd, error) {
	c := &userAddRoleCmd{userCmd: parent}
	fs, _, err := parseFlags("add-role", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Role, "role", "", "role name")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userAddRoleCmd) Run() error {
	if c.Username == "" || c.Role == "" {
		return fmt.Errorf("username and role required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	c.rootCmd.Verbosef("adding role %s to %s", c.Role, c.Username)
	u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if _, err := queries.UserHasRole(ctx, db.UserHasRoleParams{UsersIdusers: u.Idusers, Name: c.Role}); err == nil {
		c.rootCmd.Verbosef("%s already has role %s", c.Username, c.Role)
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check role: %w", err)
	}
	if err := queries.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         c.Role,
	}); err != nil {
		return fmt.Errorf("add role: %w", err)
	}
	c.rootCmd.Infof("added role %s to %s", c.Role, c.Username)
	return nil
}
