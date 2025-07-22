package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userRemoveRoleCmd implements the "user remove-role" command.
type userRemoveRoleCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	Role     string
	args     []string
}

func parseUserRemoveRoleCmd(parent *userCmd, args []string) (*userRemoveRoleCmd, error) {
	c := &userRemoveRoleCmd{userCmd: parent}
	fs, rest, err := parseFlags("remove-role", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Role, "role", "", "role name")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = rest
	return c, nil
}

func (c *userRemoveRoleCmd) Run() error {
	if c.Username == "" || c.Role == "" {
		return fmt.Errorf("username and role required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	c.rootCmd.Verbosef("removing role %s from %s", c.Role, c.Username)
	u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	perms, err := queries.GetPermissionsByUserID(ctx, u.Idusers)
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}
	for _, p := range perms {
		if p.Name == c.Role {
			if err := queries.DeleteUserRole(ctx, p.IduserRoles); err != nil {
				return fmt.Errorf("remove role: %w", err)
			}
			c.rootCmd.Infof("removed role %s from %s", c.Role, c.Username)
			return nil
		}
	}
	return fmt.Errorf("role %s not found for %s", c.Role, c.Username)
}
