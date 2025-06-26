package main

import (
	"flag"
	"fmt"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s email queue <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  list\tlist queued emails")
	fmt.Fprintln(w, "  resend\tresend a queued email")
	fmt.Fprintln(w, "  delete\tdelete a queued email")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s email queue list\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s email queue resend -id 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s email queue delete -id 1\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
