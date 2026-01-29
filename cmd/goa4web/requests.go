package main

import (
	"flag"
	"fmt"
)

// requestsCmd handles request queue management subcommands.
type requestsCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseRequestsCmd(parent *rootCmd, args []string) (*requestsCmd, error) {
	c := &requestsCmd{rootCmd: parent}
	c.fs = newFlagSet("requests")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *requestsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing requests command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseRequestsListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "view":
		cmd, err := parseRequestsViewCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("view: %w", err)
    }
	case "accept":
		cmd, err := parseRequestsAcceptCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}
		return cmd.Run()
	case "reject":
		cmd, err := parseRequestsRejectCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("reject: %w", err)
		}
		return cmd.Run()
	case "comment":
		cmd, err := parseRequestsCommentCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("comment: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown requests command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *requestsCmd) Usage() {
	executeUsage(c.fs.Output(), "requests_usage.txt", c)
}

func (c *requestsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*requestsCmd)(nil)
