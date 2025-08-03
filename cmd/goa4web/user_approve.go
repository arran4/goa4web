package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userApproveCmd approves a pending user account.
type userApproveCmd struct {
	*userCmd
	fs       *flag.FlagSet
	ID       int
	Username string
}

func parseUserApproveCmd(parent *userCmd, args []string) (*userApproveCmd, error) {
	c := &userApproveCmd{userCmd: parent}
	fs, _, err := parseFlags("approve", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userApproveCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if c.ID == 0 {
		u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	c.rootCmd.Verbosef("approving user %d", c.ID)
	if err := queries.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{UsersIdusers: int32(c.ID), Name: "user"}); err != nil {
		return fmt.Errorf("add role: %w", err)
	}
	c.rootCmd.Infof("approved user %d", c.ID)
	return nil
}
