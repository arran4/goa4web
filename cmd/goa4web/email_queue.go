package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/email_queue_usage.txt
var emailQueueUsageTemplate string

// emailQueueCmd handles email queue operations.
type emailQueueCmd struct {
	*emailCmd
	fs   *flag.FlagSet
	args []string
}

func parseEmailQueueCmd(parent *emailCmd, args []string) (*emailQueueCmd, error) {
	c := &emailQueueCmd{emailCmd: parent}
	fs := flag.NewFlagSet("queue", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *emailQueueCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing queue command")
	}
	switch c.args[0] {
	case "list":
		cmd, err := parseEmailQueueListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "resend":
		cmd, err := parseEmailQueueResendCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("resend: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseEmailQueueDeleteCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown queue command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *emailQueueCmd) Usage() {
	executeUsage(c.fs.Output(), emailQueueUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
