package main

import (
	"flag"
	"fmt"
)

// notificationsCmd handles notification-related subcommands.
type notificationsCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseNotificationsCmd(parent *rootCmd, args []string) (*notificationsCmd, error) {
	c := &notificationsCmd{rootCmd: parent}
	c.fs = newFlagSet("notifications")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *notificationsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing notifications command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "tasks":
		cmd, err := parseNotificationsTasksCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("tasks: %w", err)
		}
		return cmd.Run()
	case "send":
		cmd, err := parseNotificationsSendCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("send: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown notifications command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *notificationsCmd) Usage() {
	executeUsage(c.fs.Output(), "notifications_usage.txt", c)
}

func (c *notificationsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*notificationsCmd)(nil)
