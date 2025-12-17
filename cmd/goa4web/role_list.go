package main

import (
	"flag"
	"fmt"
)

// roleListCmd implements the "role list" subcommand.
type roleListCmd struct {
	*roleCmd
	fs *flag.FlagSet
}

// TODO make it clear that this is listing sql statements and that the actual role name is different (has spaces for one.) Should probably be 2 separate commands.

func parseRoleListCmd(parent *roleCmd, args []string) (*roleListCmd, error) {
	c := &roleListCmd{roleCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *roleListCmd) Run() error {
	roles, err := listEmbeddedRoles()
	if err != nil {
		return err
	}
	for _, r := range roles {
		fmt.Fprintln(c.fs.Output(), r)
	}
	return nil
}

func (c *roleListCmd) Usage() { executeUsage(c.fs.Output(), "role_list_usage.txt", c) }

func (c *roleListCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleListCmd)(nil)
