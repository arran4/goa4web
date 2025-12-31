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
	dryRun    bool
}

func parseUserExpungeUnverifiedCmd(parent *userCmd, args []string) (*userExpungeUnverifiedCmd, error) {
	c := &userExpungeUnverifiedCmd{userCmd: parent}
	c.fs = newFlagSet("expunge-unverified")
	c.fs.DurationVar(&c.olderThan, "older-than", 0, "Duration to define 'older than' (e.g., 72h).")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "List emails that would be deleted without taking action.")
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

	cutoff := time.Now().Add(-c.olderThan)

	if c.dryRun {
		// We need a list query for this
		// SystemListUnverifiedEmailsExpiresBefore(cutoff)
		// I need to add this query or reuse existing
		// SystemDeleteUnverifiedEmailsExpiresBefore uses `verification_expires_at < ?`
		// I should check `internal/db/queries-user_emails.sql` if there is a list variant.
		// I don't think I added one. I should add `SystemListUnverifiedEmailsExpiresBefore`
		return fmt.Errorf("dry-run not implemented yet, missing query")
	}

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
