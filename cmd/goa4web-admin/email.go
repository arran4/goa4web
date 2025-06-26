package main

import (
	"flag"
	"fmt"
)

// emailCmd handles email-related subcommands.
type emailCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseEmailCmd(parent *rootCmd, args []string) (*emailCmd, error) {
	c := &emailCmd{rootCmd: parent}
	fs := flag.NewFlagSet("email", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *emailCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing email command")
	}
	switch c.args[0] {
	case "queue":
		cmd, err := parseEmailQueueCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("queue: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown email command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *emailCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s email <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  queue\tmanage queued emails")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s email queue list\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
