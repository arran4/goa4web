package main

import (
	"flag"
	"fmt"
)

// emailQueueCmd handles email queue operations.
type emailQueueCmd struct {
	*emailCmd
	fs *flag.FlagSet
}

func parseEmailQueueCmd(parent *emailCmd, args []string) (*emailQueueCmd, error) {
	c := &emailQueueCmd{emailCmd: parent}
	c.fs = newFlagSet("queue")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailQueueCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing queue command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseEmailQueueListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "resend":
		cmd, err := parseEmailQueueResendCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("resend: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseEmailQueueDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown queue command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *emailQueueCmd) Usage() {
	executeUsage(c.fs.Output(), "email_queue_usage.txt", c)
}

func (c *emailQueueCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*emailQueueCmd)(nil)
