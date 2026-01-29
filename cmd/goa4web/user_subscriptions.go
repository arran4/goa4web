package main

import (
	"flag"
	"fmt"
)

// userSubscriptionsCmd handles "user subscriptions".
type userSubscriptionsCmd struct {
	*userCmd
	fs *flag.FlagSet
}

func parseUserSubscriptionsCmd(parent *userCmd, args []string) (*userSubscriptionsCmd, error) {
	c := &userSubscriptionsCmd{userCmd: parent}
	c.fs = newFlagSet("subscriptions")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userSubscriptionsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.Usage()
		return fmt.Errorf("missing subscriptions command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseUserSubscriptionsListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	default:
		c.Usage()
		return fmt.Errorf("unknown subscriptions command %q", args[0])
	}
}

func (c *userSubscriptionsCmd) Usage() {
	executeUsage(c.fs.Output(), "user_subscriptions_usage.txt", c)
}

func (c *userSubscriptionsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userSubscriptionsCmd)(nil)
