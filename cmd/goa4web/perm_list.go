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
	args []string
}

func parsePermListCmd(parent *permCmd, args []string) (*permListCmd, error) {
	c := &permListCmd{permCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.StringVar(&c.User, "user", "", "username filter")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
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
