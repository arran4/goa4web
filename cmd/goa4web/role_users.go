package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

//go:embed templates/role_users_usage.txt
var roleUsersUsageTemplate string

// roleUsersCmd implements "role users".
type roleUsersCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	args []string
}

func parseRoleUsersCmd(parent *roleCmd, args []string) (*roleUsersCmd, error) {
	c := &roleUsersCmd{roleCmd: parent}
	fs := flag.NewFlagSet("users", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *roleUsersCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.ListRolesWithUsers(ctx)
	if err != nil {
		return fmt.Errorf("list roles with users: %w", err)
	}
	for _, r := range rows {
		userList := ""
		if r.Users.Valid {
			userList = r.Users.String
		}
		fmt.Printf("%s: %s\n", r.Name, userList)
	}
	return nil
}

func (c *roleUsersCmd) Usage() {
	executeUsage(c.fs.Output(), roleUsersUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
