package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// ipBanUpdateCmd implements "ipban update".
type ipBanUpdateCmd struct {
	*ipBanCmd
	fs      *flag.FlagSet
	ID      int
	Reason  string
	Expires string
}

func parseIpBanUpdateCmd(parent *ipBanCmd, args []string) (*ipBanUpdateCmd, error) {
	c := &ipBanUpdateCmd{ipBanCmd: parent}
	fs, _, err := parseFlags("update", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "ban id")
		fs.StringVar(&c.Reason, "reason", "", "ban reason")
		fs.StringVar(&c.Expires, "expires", "", "expiry date YYYY-MM-DD")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *ipBanUpdateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	var expires sql.NullTime
	if c.Expires != "" {
		t, err := time.Parse("2006-01-02", c.Expires)
		if err != nil {
			return fmt.Errorf("parse expires: %w", err)
		}
		expires = sql.NullTime{Time: t, Valid: true}
	}
	err = queries.AdminUpdateBannedIp(ctx, dbpkg.AdminUpdateBannedIpParams{
		Reason:    sql.NullString{String: c.Reason, Valid: c.Reason != ""},
		ExpiresAt: expires,
		ID:        int32(c.ID),
	})
	if err != nil {
		return fmt.Errorf("update ban: %w", err)
	}
	return nil
}
