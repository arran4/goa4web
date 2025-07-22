package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// permListCmd implements "perm list".
type permListCmd struct {
	*permCmd
	fs   *flag.FlagSet
	User string
}

func parsePermListCmd(parent *permCmd, args []string) (*permListCmd, error) {
	c := &permListCmd{permCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.StringVar(&c.User, "user", "", "username filter")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *permListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.GetPermissionsWithUsers(ctx,
		dbpkg.GetPermissionsWithUsersParams{Username: sql.NullString{String: c.User, Valid: c.User != ""}},
	)
	if err != nil {
		return fmt.Errorf("list permissions: %w", err)
	}
	for _, p := range rows {
		fmt.Printf("%d\t%s\t%s\n", p.IduserRoles, p.Username.String, p.Name)
	}
	return nil
}
