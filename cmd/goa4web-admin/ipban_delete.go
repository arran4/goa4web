package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// ipBanDeleteCmd implements "ipban delete".
type ipBanDeleteCmd struct {
	*ipBanCmd
	fs   *flag.FlagSet
	IP   string
	args []string
}

func parseIpBanDeleteCmd(parent *ipBanCmd, args []string) (*ipBanDeleteCmd, error) {
	c := &ipBanDeleteCmd{ipBanCmd: parent}
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.StringVar(&c.IP, "ip", "", "ip or cidr")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *ipBanDeleteCmd) Run() error {
	if c.IP == "" {
		return fmt.Errorf("ip required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if err := queries.CancelBannedIp(ctx, c.IP); err != nil {
		return fmt.Errorf("cancel banned ip: %w", err)
	}
	return nil
}
