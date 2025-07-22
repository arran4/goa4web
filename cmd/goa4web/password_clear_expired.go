package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// passwordClearExpiredCmd implements "password clear-expired".
type passwordClearExpiredCmd struct {
	*passwordCmd
	fs    *flag.FlagSet
	Hours int
	args  []string
}

func parsePasswordClearExpiredCmd(parent *passwordCmd, args []string) (*passwordClearExpiredCmd, error) {
	c := &passwordClearExpiredCmd{passwordCmd: parent, Hours: 24}
	fs := flag.NewFlagSet("clear-expired", flag.ContinueOnError)
	fs.IntVar(&c.Hours, "hours", 24, "expiration age in hours")
	c.fs = fs
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *passwordClearExpiredCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	expiry := time.Now().Add(-time.Duration(c.Hours) * time.Hour)
	if err := queries.PurgePasswordResetsBefore(ctx, expiry); err != nil {
		return fmt.Errorf("clear expired: %w", err)
	}
	return nil
}
