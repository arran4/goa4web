package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/db"
)

// userRejectCmd rejects a pending user account.
type userRejectCmd struct {
	*userCmd
	fs       *flag.FlagSet
	ID       int
	Username string
	Reason   string
}

func parseUserRejectCmd(parent *userCmd, args []string) (*userRejectCmd, error) {
	c := &userRejectCmd{userCmd: parent}
	fs, _, err := parseFlags("reject", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Reason, "reason", "", "rejection reason")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userRejectCmd) Usage() {
	executeUsage(c.fs.Output(), "user_reject_usage.txt", c)
}

func (c *userRejectCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userRejectCmd)(nil)

func (c *userRejectCmd) Run() error {
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
	c.rootCmd.Verbosef("rejecting user %d", c.ID)
	if err := queries.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{UsersIdusers: int32(c.ID), Name: "rejected"}); err != nil {
		return fmt.Errorf("add role: %w", err)
	}
	if c.Reason != "" {
		if err := queries.InsertAdminUserComment(ctx, db.InsertAdminUserCommentParams{UsersIdusers: int32(c.ID), Comment: c.Reason}); err != nil {
			log.Printf("insert admin user comment: %v", err)
		}
	}
	c.rootCmd.Infof("rejected user %d", c.ID)
	return nil
}
