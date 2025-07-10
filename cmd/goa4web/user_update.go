package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userUpdateCmd implements the "user update" command.
type userUpdateCmd struct {
	*userCmd
	fs         *flag.FlagSet
	Username   string
	Email      string
	MakeAdmin  bool
	MakeWriter bool
	args       []string
}

func parseUserUpdateCmd(parent *userCmd, args []string) (*userUpdateCmd, error) {
	c := &userUpdateCmd{userCmd: parent}
	fs, rest, err := parseFlags("update", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Email, "email", "", "email address")
		fs.BoolVar(&c.MakeAdmin, "make-admin", false, "set role to administrator")
		fs.BoolVar(&c.MakeWriter, "make-writer", false, "set role to writer")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = rest
	return c, nil
}

func (c *userUpdateCmd) Run() error {
	if c.Username == "" {
		return fmt.Errorf("username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if c.Email != "" {
		if err := queries.UpdateUserEmail(ctx, dbpkg.UpdateUserEmailParams{
			Email:  c.Email,
			UserID: u.Idusers,
		}); err != nil {
			return fmt.Errorf("update email: %w", err)
		}
	}
	if c.MakeAdmin || c.MakeWriter {
		perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(ctx,
			dbpkg.GetPermissionsByUserIdAndSectionAndSectionAllParams{
				UsersIdusers: u.Idusers,
				Section:      sql.NullString{String: "all", Valid: true},
			})
		if err == nil && perm != nil {
			if c.MakeWriter && perm.Level.String == "administrator" && c.rootCmd.Verbosity > 0 {
				fmt.Printf("warning: removing administrator from %s\n", c.Username)
			}
			if err := queries.PermissionUserDisallow(ctx, perm.Idpermissions); err != nil {
				return fmt.Errorf("update role: %w", err)
			}
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check role: %w", err)
		}
		level := "writer"
		if c.MakeAdmin {
			level = "administrator"
		}
		if err := queries.PermissionUserAllow(ctx, dbpkg.PermissionUserAllowParams{
			UsersIdusers: u.Idusers,
			Section:      sql.NullString{String: "all", Valid: true},
			Level:        sql.NullString{String: level, Valid: true},
		}); err != nil {
			return fmt.Errorf("set role: %w", err)
		}
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("updated user %s\n", c.Username)
	}
	return nil
}
