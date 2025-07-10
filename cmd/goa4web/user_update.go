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
	fs          *flag.FlagSet
	Username    string
	Email       string
	MakeAdmin   bool
	RemoveAdmin bool
	args        []string
}

func parseUserUpdateCmd(parent *userCmd, args []string) (*userUpdateCmd, error) {
	c := &userUpdateCmd{userCmd: parent}
	fs, rest, err := parseFlags("update", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Email, "email", "", "email address")
		fs.BoolVar(&c.MakeAdmin, "make-admin", false, "grant administrator rights")
		fs.BoolVar(&c.RemoveAdmin, "remove-admin", false, "revoke administrator rights")
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
	if c.MakeAdmin {
		if _, err := queries.GetAdministratorPermissionByUserId(ctx, u.Idusers); err == nil {
			if c.rootCmd.Verbosity > 0 {
				fmt.Printf("%s already administrator\n", c.Username)
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check admin: %w", err)
		} else if err := queries.PermissionUserAllow(ctx, dbpkg.PermissionUserAllowParams{
			UsersIdusers: u.Idusers,
			Section:      sql.NullString{String: "all", Valid: true},
			Level:        sql.NullString{String: "administrator", Valid: true},
		}); err != nil {
			return fmt.Errorf("make admin: %w", err)
		}
	}
	if c.RemoveAdmin {
		perm, err := queries.GetUserPermissions(ctx, u.Idusers)
		if err == nil && perm != nil {
			if err := queries.PermissionUserDisallow(ctx, perm.Idpermissions); err != nil {
				return fmt.Errorf("remove admin: %w", err)
			}
		}
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("updated user %s\n", c.Username)
	}
	return nil
}
