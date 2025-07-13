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
	fs       *flag.FlagSet
	Username string
	Email    string
	Role     string
	args     []string
}

func parseUserUpdateCmd(parent *userCmd, args []string) (*userUpdateCmd, error) {
	c := &userUpdateCmd{userCmd: parent}
	fs, rest, err := parseFlags("update", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Email, "email", "", "email address")
		fs.StringVar(&c.Role, "role", "", "set user role (administrator, user, content writer, anonymous)")
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
	if c.Role != "" {
		switch c.Role {
		case "administrator", "user", "content writer", "moderator", "anonymous":
		default:
			return fmt.Errorf("invalid role %q", c.Role)
		}

		perms, err := queries.GetPermissionsByUserID(ctx, u.Idusers)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check role: %w", err)
		}

		var perm *dbpkg.GetUserRolesRow
		if len(perms) > 0 {
			perm = perms[0]
		}

		if perm != nil && perm.Role != "" {
			if perm.Role == "administrator" && c.Role != "administrator" && c.rootCmd.Verbosity > 0 {
				fmt.Printf("warning: removing administrator from %s\n", c.Username)
			}
			if c.Role == "anonymous" || perm.Role != c.Role {
				if err := queries.DeleteUserRole(ctx, perm.IduserRoles); err != nil {
					return fmt.Errorf("update role: %w", err)
				}
				perm = nil
			}
		}

		if c.Role != "anonymous" && (perm == nil || perm.Role != c.Role) {
			if err := queries.CreateUserRole(ctx, dbpkg.CreateUserRoleParams{
				UsersIdusers: u.Idusers,
				Name:         c.Role,
			}); err != nil {
				return fmt.Errorf("set role: %w", err)
			}
		}
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("updated user %s\n", c.Username)
	}
	return nil
}
