package main

import (
	"context"

	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userRolesCmd implements "user roles".
type userRolesCmd struct {
	*userCmd
	fs *flag.FlagSet
}

func parseUserRolesCmd(parent *userCmd, args []string) (*userRolesCmd, error) {
	c := &userRolesCmd{userCmd: parent}
	c.fs = newFlagSet("roles")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userRolesCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	rows, err := queries.ListUsersWithRoles(ctx)
	if err != nil {
		return fmt.Errorf("list users with roles: %w", err)
	}
	for _, r := range rows {
		roleList := ""
		if r.Roles.Valid {
			roleList = r.Roles.String
		}
		fmt.Printf("%s\t%s\n", r.Username.String, roleList)
	}
	return nil
}

func (c *userRolesCmd) Usage() {
	executeUsage(c.fs.Output(), "user_roles_usage.txt", c)
}

func (c *userRolesCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userRolesCmd)(nil)
