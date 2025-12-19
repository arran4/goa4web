package main

import (
	"flag"
	"fmt"
)

// roleListCmd implements the "role list" subcommand.
type roleListCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	args []string
}

func parseRoleListCmd(parent *roleCmd, args []string) (*roleListCmd, error) {
	c := &roleListCmd{roleCmd: parent}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *roleListCmd) Run() error {
	if len(c.args) == 0 {
		cmd, err := parseRoleListAllCmd(c, nil)
		if err != nil {
			return fmt.Errorf("all: %w", err)
		}
		return cmd.Run()
	}
	if err := usageIfHelp(c.fs, c.args); err != nil {
		return err
	}
	switch c.args[0] {
	case "sql":
		cmd, err := parseRoleListSQLCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("sql: %w", err)
		}
		return cmd.Run()
	case "names":
		cmd, err := parseRoleListNamesCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("names: %w", err)
		}
		return cmd.Run()
	case "all":
		cmd, err := parseRoleListAllCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("all: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown list command %q", c.args[0])
	}
}

func (c *roleListCmd) Usage() { executeUsage(c.fs.Output(), "role_list_usage.txt", c) }

func (c *roleListCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleListCmd)(nil)

type roleListSQLCmd struct {
	*roleListCmd
	fs *flag.FlagSet
}

func parseRoleListSQLCmd(parent *roleListCmd, args []string) (*roleListSQLCmd, error) {
	c := &roleListSQLCmd{roleListCmd: parent}
	fs := flag.NewFlagSet("sql", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *roleListSQLCmd) Run() error {
	roles, err := listEmbeddedRoles()
	if err != nil {
		return err
	}
	for _, r := range roles {
		fmt.Fprintln(c.fs.Output(), r)
	}
	return nil
}

func (c *roleListSQLCmd) Usage() { executeUsage(c.fs.Output(), "role_list_sql_usage.txt", c) }

func (c *roleListSQLCmd) FlagGroups() []flagGroup {
	return append(c.roleCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

type roleListNamesCmd struct {
	*roleListCmd
	fs *flag.FlagSet
}

func parseRoleListNamesCmd(parent *roleListCmd, args []string) (*roleListNamesCmd, error) {
	c := &roleListNamesCmd{roleListCmd: parent}
	fs := flag.NewFlagSet("names", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *roleListNamesCmd) Run() error {
	names, err := listEmbeddedRoleNames()
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Fprintln(c.fs.Output(), name)
	}
	return nil
}

func (c *roleListNamesCmd) Usage() { executeUsage(c.fs.Output(), "role_list_names_usage.txt", c) }

func (c *roleListNamesCmd) FlagGroups() []flagGroup {
	return append(c.roleCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleListSQLCmd)(nil)
var _ usageData = (*roleListNamesCmd)(nil)
