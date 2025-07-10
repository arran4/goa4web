package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// permGrantCmd implements "perm grant".
type permGrantCmd struct {
	*permCmd
	fs      *flag.FlagSet
	User    string
	Section string
	Role    string
	args    []string
}

func parsePermGrantCmd(parent *permCmd, args []string) (*permGrantCmd, error) {
	c := &permGrantCmd{permCmd: parent}
	fs := flag.NewFlagSet("grant", flag.ContinueOnError)
	fs.StringVar(&c.User, "user", "", "username")
	fs.StringVar(&c.Section, "section", "", "permission section")
	fs.StringVar(&c.Role, "role", "", "permission role")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *permGrantCmd) Run() error {
	if c.User == "" || c.Section == "" || c.Role == "" {
		return fmt.Errorf("user, section and role required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.User, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if c.Section == "all" && c.Role == "administrator" {
		if _, err := queries.GetAdministratorPermissionByUserId(ctx, u.Idusers); err == nil {
			if c.rootCmd.Verbosity > 0 {
				fmt.Printf("%s already administrator\n", c.User)
			}
			return nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check admin: %w", err)
		}
	}
	if err := queries.PermissionUserAllow(ctx, dbpkg.PermissionUserAllowParams{
		UsersIdusers: u.Idusers,
		Section:      sql.NullString{String: c.Section, Valid: true},
		Role:         sql.NullString{String: c.Role, Valid: true},
	}); err != nil {
		return fmt.Errorf("grant: %w", err)
	}
	return nil
}
