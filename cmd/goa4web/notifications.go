package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/notifications_usage.txt
var notificationsUsageTemplate string

// notificationsCmd handles notification-related subcommands.
type notificationsCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseNotificationsCmd(parent *rootCmd, args []string) (*notificationsCmd, error) {
	c := &notificationsCmd{rootCmd: parent}
	fs := flag.NewFlagSet("notifications", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *notificationsCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing notifications command")
	}
	switch c.args[0] {
	case "tasks":
		cmd, err := parseNotificationsTasksCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("tasks: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown notifications command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *notificationsCmd) Usage() {
	executeUsage(c.fs.Output(), notificationsUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
