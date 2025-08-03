package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// ipBanDeleteCmd implements "ipban delete".
type ipBanDeleteCmd struct {
	*ipBanCmd
	fs *flag.FlagSet
	IP string
}

func parseIpBanDeleteCmd(parent *ipBanCmd, args []string) (*ipBanDeleteCmd, error) {
	c := &ipBanDeleteCmd{ipBanCmd: parent}
	fs, _, err := parseFlags("delete", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.IP, "ip", "", "ip or cidr")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *ipBanDeleteCmd) Run() error {
	if c.IP == "" {
		return fmt.Errorf("ip required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if err := queries.AdminCancelBannedIp(ctx, c.IP); err != nil {
		return fmt.Errorf("cancel banned ip: %w", err)
	}
	return nil
}
