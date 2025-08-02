package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
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

func (c *userListRolesCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	roles, err := queries.AdminListRoles(ctx)
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}
	for _, r := range roles {
		fmt.Printf("%d\t%s\n", r.ID, r.Name)
	}
	return nil
}
