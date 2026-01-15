package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// userPasswordApproveCmd approves a pending user password change.
type userPasswordApproveCmd struct {
	*userPasswordCmd
	fs       *flag.FlagSet
	ID       int
	Username string
}

func parseUserPasswordApproveCmd(parent *userPasswordCmd, args []string) (*userPasswordApproveCmd, error) {
	c := &userPasswordApproveCmd{userPasswordCmd: parent}
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

func (c *userPasswordApproveCmd) Run() error {
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

	c.rootCmd.Verbosef("approving password for user %d", c.ID)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

	pendingPassword, err := qtx.GetPendingPassword(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("get pending password: %w", err)
	}

	if err := qtx.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		UsersIdusers:    int32(c.ID),
		Passwd:          pendingPassword.Passwd,
		PasswdAlgorithm: sql.NullString{String: pendingPassword.PasswdAlgorithm, Valid: true},
	}); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	if err := qtx.DeletePendingPassword(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("delete pending password: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	c.rootCmd.Infof("approved password for user %d", c.ID)
	return nil
}

func (c *userPasswordApproveCmd) Usage() {
	executeUsage(c.fs.Output(), "user_password_approve_usage.txt", c)
}

func (c *userPasswordApproveCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userPasswordApproveCmd)(nil)
