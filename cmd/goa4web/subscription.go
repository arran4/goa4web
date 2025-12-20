package main

import (
	"flag"
	"fmt"
)

// subscriptionCmd implements the "subscription" command.
type subscriptionCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseSubscriptionCmd(parent *rootCmd, args []string) (*subscriptionCmd, error) {
	c := &subscriptionCmd{rootCmd: parent}
	fs := flag.NewFlagSet("subscription", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *subscriptionCmd) Run() error {
	if c.fs.NArg() == 0 {
		c.Usage()
		return fmt.Errorf("missing subcommand")
	}

	switch c.fs.Arg(0) {
	case "template":
		cmd, err := parseSubscriptionTemplateCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	default:
		c.Usage()
		return fmt.Errorf("unknown subcommand: %s", c.fs.Arg(0))
	}
}

func (c *subscriptionCmd) Usage() {
	executeUsage(c.fs.Output(), "subscription_usage.txt", c)
}

func (c *subscriptionCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*subscriptionCmd)(nil)
