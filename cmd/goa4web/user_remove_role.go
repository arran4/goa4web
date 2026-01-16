package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userRemoveRoleCmd implements the "user remove-role" command.
type userRemoveRoleCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	Role     string
}

func parseUserRemoveRoleCmd(parent *userCmd, args []string) (*userRemoveRoleCmd, error) {
	c := &userRemoveRoleCmd{userCmd: parent}
	fs, _, err := parseFlags("remove-role", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Role, "role", "", "role name")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userRemoveRoleCmd) Usage() {
	executeUsage(c.fs.Output(), "user_remove_role_usage.txt", c)
}

func (c *userRemoveRoleCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userRemoveRoleCmd)(nil)

func (c *userRemoveRoleCmd) Run() error {
	if c.Username == "" || c.Role == "" {
		return fmt.Errorf("username and role required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	c.rootCmd.Verbosef("removing role %s from %s", c.Role, c.Username)
	u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	perms, err := queries.GetPermissionsByUserID(ctx, u.Idusers)
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}
	for _, p := range perms {
		if p.Name == c.Role {
			if err := queries.AdminDeleteUserRole(ctx, p.IduserRoles); err != nil {
				return fmt.Errorf("remove role: %w", err)
			}
			c.rootCmd.Infof("removed role %s from %s", c.Role, c.Username)
			return nil
		}
	}
	return fmt.Errorf("role %s not found for %s", c.Role, c.Username)
}
