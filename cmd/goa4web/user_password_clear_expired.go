package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userPasswordClearExpiredCmd implements "user password clear-expired".
// It removes expired entries from the reset password queue.
type userPasswordClearExpiredCmd struct {
	*userPasswordCmd
	fs    *flag.FlagSet
	Hours int
	List  bool
	args  []string
}

func parseUserPasswordClearExpiredCmd(parent *userPasswordCmd, args []string) (*userPasswordClearExpiredCmd, error) {
	c := &userPasswordClearExpiredCmd{userPasswordCmd: parent, Hours: 24}
	fs := flag.NewFlagSet("clear-expired", flag.ContinueOnError)
	fs.IntVar(&c.Hours, "hours", 24, "expiration age in hours")
	fs.BoolVar(&c.List, "list", false, "list removed reset requests")
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *userPasswordClearExpiredCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	expiry := time.Now().Add(-time.Duration(c.Hours) * time.Hour)
	var details []*dbpkg.ListPasswordResetsBeforeRow
	if c.List {
		var err error
		details, err = queries.ListPasswordResetsBefore(ctx, expiry)
		if err != nil {
			return fmt.Errorf("list resets: %w", err)
		}
	}
	res, err := queries.PurgePasswordResetsBefore(ctx, expiry)
	if err != nil {
		return fmt.Errorf("clear expired: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil {
		c.rootCmd.Infof("deleted %d expired password reset requests", rows)
	}
	if c.List {
		for _, r := range details {
			c.rootCmd.Infof("removed id=%d code=%s created=%s", r.ID, r.VerificationCode, r.CreatedAt.Format(time.RFC3339))
		}
	}
	return nil
}
