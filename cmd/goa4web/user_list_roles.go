package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userListRolesCmd implements the "user list-roles" command.
type userListRolesCmd struct {
	*userCmd
	fs *flag.FlagSet
}

func parseUserListRolesCmd(parent *userCmd, args []string) (*userListRolesCmd, error) {
	c := &userListRolesCmd{userCmd: parent}
	fs, _, err := parseFlags("list-roles", args, nil)
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userListRolesCmd) Usage() {
	executeUsage(c.fs.Output(), "user_list_roles_usage.txt", c)
}

func (c *userListRolesCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userListRolesCmd)(nil)

func (c *userListRolesCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	roles, err := queries.AdminListRoles(ctx)
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}
	for _, r := range roles {
		fmt.Printf("%d\t%s\n", r.ID, r.Name)
	}
	return nil
}
