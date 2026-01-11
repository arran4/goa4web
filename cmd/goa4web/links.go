package main

import (
	"flag"
	"fmt"
)

// linksCmd provides utilities for signing and verifying external links.
type linksCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseLinksCmd(parent *rootCmd, args []string) (*linksCmd, error) {
	c := &linksCmd{rootCmd: parent}
	c.fs = newFlagSet("links")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *linksCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing links command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "sign":
		cmd, err := parseLinksSignCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("sign: %w", err)
		}
		return cmd.Run()
	case "verify":
		cmd, err := parseLinksVerifyCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("verify: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseLinksListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseLinksDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "refresh":
		cmd, err := parseLinksRefreshCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("refresh: %w", err)
		}
		return cmd.Run()
	case "remap":
		cmd, err := parseLinksRemapCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("remap: %w", err)
		}
		return cmd.Run()
	case "deactivate":
		cmd, err := parseLinksDeactivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("deactivate: %w", err)
		}
		return cmd.Run()
	case "activate":
		cmd, err := parseLinksActivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("activate: %w", err)
		}
		return cmd.Run()
	case "list-deactivated":
		cmd, err := parseLinksListDeactivatedCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list-deactivated: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown links command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *linksCmd) Usage() {
	executeUsage(c.fs.Output(), "links_usage.txt", c)
}

func (c *linksCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*linksCmd)(nil)
