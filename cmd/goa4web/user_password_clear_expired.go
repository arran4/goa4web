package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// userPasswordClearExpiredCmd implements "user password clear-expired".
// It removes expired entries from the reset password queue.
type userPasswordClearExpiredCmd struct {
	*userPasswordCmd
	fs    *flag.FlagSet
	Hours int
	args  []string
}

func parseUserPasswordClearExpiredCmd(parent *userPasswordCmd, args []string) (*userPasswordClearExpiredCmd, error) {
	c := &userPasswordClearExpiredCmd{userPasswordCmd: parent, Hours: 24}
	fs := flag.NewFlagSet("clear-expired", flag.ContinueOnError)
	fs.IntVar(&c.Hours, "hours", 24, "expiration age in hours")
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *userPasswordClearExpiredCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	expiry := time.Now().Add(-time.Duration(c.Hours) * time.Hour)
	res, err := queries.SystemPurgePasswordResetsBefore(ctx, expiry)
	if err != nil {
		return fmt.Errorf("clear expired: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil {
		c.rootCmd.Infof("deleted %d expired password reset requests", rows)
	}
	return nil
}
