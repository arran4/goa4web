package main

import (
	"flag"
	"fmt"
)

// emailCmd handles email-related subcommands.
type emailCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseEmailCmd(parent *rootCmd, args []string) (*emailCmd, error) {
	c := &emailCmd{rootCmd: parent}
	c.fs = newFlagSet("email")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing email command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "failed":
		cmd, err := parseEmailFailedCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("failed: %w", err)
		}
		return cmd.Run()
	case "queue":
		cmd, err := parseEmailQueueCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("queue: %w", err)
		}
		return cmd.Run()
	case "send":
		cmd, err := parseEmailSendCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("send: %w", err)
		}
		return cmd.Run()
	case "test":
		cmd, err := parseEmailTestCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("test: %w", err)
		}
		return cmd.Run()
	case "template":
		cmd, err := parseEmailTemplateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("template: %w", err)
		}
		return cmd.Run()
	case "sent":
		cmd, err := parseEmailSentCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("sent: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown email command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *emailCmd) Usage() {
	executeUsage(c.fs.Output(), "email_usage.txt", c)
}

func (c *emailCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*emailCmd)(nil)
