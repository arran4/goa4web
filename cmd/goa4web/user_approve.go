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
	c.fs = newFlagSet("approve")
	c.fs.Usage = c.Usage

	c.fs.IntVar(&c.ID, "id", 0, "user id")
	c.fs.StringVar(&c.Username, "username", "", "username")

	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	switch remaining := c.fs.Args(); len(remaining) {
	case 0:
	case 1:
		if c.Username == "" {
			c.Username = remaining[0]
		} else {
			return nil, fmt.Errorf("unexpected arguments: %v", remaining)
		}
	default:
		return nil, fmt.Errorf("unexpected arguments: %v", remaining)
	}

	return c, nil
}

func (c *userApproveCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	queries, err := c.rootCmd.Querier()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
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

func (c *userApproveCmd) Usage() {
	executeUsage(c.fs.Output(), "user_approve_usage.txt", c)
}

func (c *userApproveCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userApproveCmd)(nil)
