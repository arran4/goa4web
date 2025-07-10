package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// permUpdateCmd implements "perm update".
type permUpdateCmd struct {
	*permCmd
	fs      *flag.FlagSet
	ID      int
	Section string
	Role    string
	args    []string
}

func parsePermUpdateCmd(parent *permCmd, args []string) (*permUpdateCmd, error) {
	c := &permUpdateCmd{permCmd: parent}
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "permission id")
	fs.StringVar(&c.Section, "section", "", "permission section")
	fs.StringVar(&c.Role, "role", "", "permission role")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *permUpdateCmd) Run() error {
	if c.ID == 0 || c.Section == "" || c.Role == "" {
		return fmt.Errorf("id, section and role required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if err := queries.UpdatePermission(ctx, dbpkg.UpdatePermissionParams{
		ID:      int32(c.ID),
		Section: sql.NullString{String: c.Section, Valid: true},
		Role:    sql.NullString{String: c.Role, Valid: true},
	}); err != nil {
		return fmt.Errorf("update permission: %w", err)
	}
	return nil
}
