package main

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
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

	cfg, err := c.loadConfig()
	if err != nil {
		return err
	}
	d, err := c.rootCmd.InitDB(cfg)
	if err != nil {
		return err
	}
	defer d.Close()
	queries := db.New(d)

	cutoff := time.Now().Add(-c.olderThan)

	if c.dryRun {
		es, err := queries.SystemListUnverifiedEmailsExpiresBefore(c.rootCmd.Context(), sql.NullTime{Time: cutoff, Valid: true})
		if err != nil {
			return fmt.Errorf("list unverified emails: %w", err)
		}
		w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tUserID\tEmail\tExpires")
		for _, e := range es {
			expires := "N/A"
			if e.VerificationExpiresAt.Valid {
				expires = e.VerificationExpiresAt.Time.Format(time.RFC3339)
			}
			fmt.Fprintf(w, "%d\t%d\t%s\t%s\n", e.ID, e.UserID, e.Email, expires)
		}
		w.Flush()
		return nil
	}

	res, err := queries.SystemDeleteUnverifiedEmailsExpiresBefore(c.rootCmd.Context(), sql.NullTime{Time: cutoff, Valid: true})
	if err != nil {
		return fmt.Errorf("expunge unverified emails: %w", err)
	}

	rows, err := res.RowsAffected()
	if err !=
nil {
		c.Infof("Expunged unknown number of unverified emails")
	} else {
		c.Infof("Expunged %d unverified emails", rows)
	}
	return nil
}

func (c *userExpungeUnverifiedCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s user expunge-unverified [flags]\n", os.Args[0])
	c.fs.PrintDefaults()
}

func (c *userExpungeUnverifiedCmd) loadConfig() (*config.RuntimeConfig, error) {
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}
	return config.NewRuntimeConfig(
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	), nil
}
