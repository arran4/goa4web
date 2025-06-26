package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// ipBanListCmd implements "ipban list".
type ipBanListCmd struct {
	*ipBanCmd
	fs   *flag.FlagSet
	args []string
}

func parseIpBanListCmd(parent *ipBanCmd, args []string) (*ipBanListCmd, error) {
	c := &ipBanListCmd{ipBanCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *ipBanListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.ListBannedIps(ctx)
	if err != nil {
		return fmt.Errorf("list banned ips: %w", err)
	}
	for _, b := range rows {
		fmt.Printf("%d\t%s\t%s\t%v\t%v\n", b.ID, b.IpNet, b.Reason.String, b.CreatedAt, b.ExpiresAt)
	}
	return nil
}
