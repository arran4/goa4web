package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
)

// userApproveCmd approves a pending user account.
type userApproveCmd struct {
	*userCmd
	fs       *flag.FlagSet
	ID       int
	Username string
	args     []string
}

func parseUserApproveCmd(parent *userCmd, args []string) (*userApproveCmd, error) {
	c := &userApproveCmd{userCmd: parent}
	fs, rest, err := parseFlags("approve", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = rest
	return c, nil
}

func (c *userApproveCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if c.ID == 0 {
		u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	if err := queries.CreateUserRole(ctx, dbpkg.CreateUserRoleParams{UsersIdusers: int32(c.ID), Name: "user"}); err != nil {
		return fmt.Errorf("add role: %w", err)
	}
	if u, err := queries.GetUserById(ctx, int32(c.ID)); err == nil && u.Email.Valid {
		_ = notif.CreateEmailTemplateAndQueue(ctx, queries, int32(c.ID), u.Email.String, "", "user approved", nil)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("approved user %d\n", c.ID)
	}
	return nil
}
