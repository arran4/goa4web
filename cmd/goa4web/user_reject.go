package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
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

func (c *userRejectCmd) Run() error {
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
	c.rootCmd.Verbosef("rejecting user %d", c.ID)
	if err := queries.CreateUserRole(ctx, dbpkg.CreateUserRoleParams{UsersIdusers: int32(c.ID), Name: "rejected"}); err != nil {
		return fmt.Errorf("add role: %w", err)
	}
	if c.Reason != "" {
		_ = queries.InsertAdminUserComment(ctx, dbpkg.InsertAdminUserCommentParams{UsersIdusers: int32(c.ID), Comment: c.Reason})
	}
	c.rootCmd.Infof("rejected user %d", c.ID)
	return nil
}
