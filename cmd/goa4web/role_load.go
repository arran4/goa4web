package main

import (
	"flag"
	"fmt"
)

// roleLoadCmd implements the "role load" subcommand.
type roleLoadCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	role string
	file string
}

func parseRoleLoadCmd(parent *roleCmd, args []string) (*roleLoadCmd, error) {
	c := &roleLoadCmd{roleCmd: parent}
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.role, "role", "", "The name of the role to load.")
	fs.StringVar(&c.file, "file", "", "Optional path to a .sql file to load instead of the embedded role script.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.role == "" {
		return nil, fmt.Errorf("role name is required")
	}
	return c, nil
}

func (c *roleLoadCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	return loadRole(sdb, c.role, c.file)
}

func (c *roleLoadCmd) Usage() {
	executeUsage(c.fs.Output(), "role_load_usage.txt", c)
}

func (c *roleLoadCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleLoadCmd)(nil)
