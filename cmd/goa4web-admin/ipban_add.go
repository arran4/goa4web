package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"time"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// ipBanAddCmd implements "ipban add".
type ipBanAddCmd struct {
	*ipBanCmd
	fs      *flag.FlagSet
	IP      string
	Reason  string
	Expires string
	args    []string
}

func parseIpBanAddCmd(parent *ipBanCmd, args []string) (*ipBanAddCmd, error) {
	c := &ipBanAddCmd{ipBanCmd: parent}
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.StringVar(&c.IP, "ip", "", "ip or cidr")
	fs.StringVar(&c.Reason, "reason", "", "ban reason")
	fs.StringVar(&c.Expires, "expires", "", "expiry date YYYY-MM-DD")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *ipBanAddCmd) Run() error {
	if c.IP == "" {
		return fmt.Errorf("ip required")
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
	err = queries.InsertBannedIp(ctx, dbpkg.InsertBannedIpParams{
		IpNet:     c.IP,
		Reason:    sql.NullString{String: c.Reason, Valid: c.Reason != ""},
		ExpiresAt: expires,
	})
	if err != nil {
		return fmt.Errorf("insert banned ip: %w", err)
	}
	return nil
}
