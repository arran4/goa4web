package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/roles"
)

// roleApplyCmd implements the "role apply" subcommand.
type roleApplyCmd struct {
	*roleCmd
	fs       *flag.FlagSet
	srcRole  string
	destRole string
}

func parseRoleApplyCmd(parent *roleCmd, args []string) (*roleApplyCmd, error) {
	c := &roleApplyCmd{roleCmd: parent}
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.srcRole, "src", "", "The source role.")
	fs.StringVar(&c.destRole, "dest", "", "The destination role.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.srcRole == "" || c.destRole == "" {
		return nil, fmt.Errorf("source and destination roles are required")
	}
	return c, nil
}

func (c *roleApplyCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.New(sdb)
	return roles.ApplyRoleGrants(c.rootCmd.ctx, sdb, q, c.srcRole, c.destRole)
}

func (c *roleApplyCmd) Usage() {
	executeUsage(c.fs.Output(), "role_apply_usage.txt", c)
}

func (c *roleApplyCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleApplyCmd)(nil)
