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
	Level   string
	args    []string
}

func parsePermUpdateCmd(parent *permCmd, args []string) (*permUpdateCmd, error) {
	c := &permUpdateCmd{permCmd: parent}
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "permission id")
	fs.StringVar(&c.Section, "section", "", "permission section")
	fs.StringVar(&c.Level, "level", "", "permission level")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *permUpdateCmd) Run() error {
	if c.ID == 0 || c.Section == "" || c.Level == "" {
		return fmt.Errorf("id, section and level required")
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
		Level:   sql.NullString{String: c.Level, Valid: true},
	}); err != nil {
		return fmt.Errorf("update permission: %w", err)
	}
	return nil
}
