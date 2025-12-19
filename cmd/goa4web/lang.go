package main

import (
	"flag"
	"fmt"
)

type langCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseLangCmd(parent *rootCmd, args []string) (*langCmd, error) {
	c := &langCmd{rootCmd: parent}
	c.fs = newFlagSet("lang")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *langCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing lang command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "add":
		cmd, err := parseLangAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseLangListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseLangUpdateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown lang command %q", args[0])
	}
}

func (c *langCmd) Usage() {
	executeUsage(c.fs.Output(), "lang_usage.txt", c)
}

func (c *langCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*langCmd)(nil)
