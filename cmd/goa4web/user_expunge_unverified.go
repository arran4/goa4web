package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
)

type userExpungeUnverifiedCmd struct {
	*userCmd
	olderThan time.Duration
}

func parseUserExpungeUnverifiedCmd(parent *userCmd, args []string) (*userExpungeUnverifiedCmd, error) {
	c := &userExpungeUnverifiedCmd{userCmd: parent}
	c.fs = newFlagSet("expunge-unverified")
	c.fs.DurationVar(&c.olderThan, "older-than", 0, "Duration to define 'older than' (e.g., 72h).")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userExpungeUnverifiedCmd) Run() error {
	if c.olderThan <= 0 {
		return fmt.Errorf("missing or invalid -older-than duration")
	}

	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("load config file: %w", err)
	}
	cfg := config.NewRuntimeConfig(
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	)

	d, err := c.rootCmd.InitDB(cfg)
	if err != nil {
		return err
	}
	defer d.Close()
	queries := db.New(d)

	// "Older than X" => verification_expires_at < now - X
	// Actually, verification_expires_at is usually created + 24h.
	// So if created T, expires T+24.
	// If we want created < Now - X.
	// T < N - X
	// T + 24 < N - X + 24
	// Expires < N - X + 24.
	//
	// However, simplistically, let's just say if it expired before (Now - X), it's definitely older than X (assuming expiry > creation).
	// If expiry is 24h, and we want older than 72h.
	// We want expired before Now - 72h.
	// If it expired 73h ago, it was created 73+24 = 97h ago. Correct.
	// So `verification_expires_at < Now - olderThan`.

	cutoff := time.Now().Add(-c.olderThan)
	res, err := queries.SystemDeleteUnverifiedEmailsExpiresBefore(c.rootCmd.Context(), sql.NullTime{Time: cutoff, Valid: true})
	if err != nil {
		return fmt.Errorf("expunge unverified emails: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		c.Infof("Expunged unknown number of unverified emails")
	} else {
		c.Infof("Expunged %d unverified emails", rows)
	}

	return nil
}

func (c *userExpungeUnverifiedCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s user expunge-unverified [flags]\n\nFlags:\n", os.Args[0])
	c.fs.PrintDefaults()
}
