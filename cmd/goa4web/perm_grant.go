package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// permGrantCmd implements "perm grant".
type permGrantCmd struct {
	*permCmd
	fs   *flag.FlagSet
	User string
	Role string
	args []string
}

func parsePermGrantCmd(parent *permCmd, args []string) (*permGrantCmd, error) {
	c := &permGrantCmd{permCmd: parent}
	fs := flag.NewFlagSet("grant", flag.ContinueOnError)
	fs.StringVar(&c.User, "user", "", "username")
	fs.StringVar(&c.Role, "role", "", "permission role")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *permGrantCmd) Run() error {
	if c.User == "" || c.Role == "" {
		return fmt.Errorf("user and role required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	c.rootCmd.Verbosef("granting %s to %s", c.Role, c.User)
	u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.User, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if c.Role == "administrator" {
		if _, err := queries.GetAdministratorUserRole(ctx, u.Idusers); err == nil {
			c.rootCmd.Verbosef("%s already administrator", c.User)
			return nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check admin: %w", err)
		}
	}
	if err := queries.CreateUserRole(ctx, dbpkg.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         c.Role,
	}); err != nil {
		return fmt.Errorf("grant: %w", err)
	}
	c.rootCmd.Infof("granted %s to %s", c.Role, c.User)
	return nil
}
