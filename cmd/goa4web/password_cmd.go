package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/password_usage.txt
var passwordUsageTemplate string

// passwordCmd handles pending password operations.
type passwordCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parsePasswordCmd(parent *rootCmd, args []string) (*passwordCmd, error) {
	c := &passwordCmd{rootCmd: parent}
	fs := flag.NewFlagSet("password", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *passwordCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing password command")
	}
	switch c.args[0] {
	case "clear-expired":
		cmd, err := parsePasswordClearExpiredCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("clear-expired: %w", err)
		}
		return cmd.Run()
	case "clear-user":
		cmd, err := parsePasswordClearUserCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("clear-user: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown password command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *passwordCmd) Usage() {
	executeUsage(c.fs.Output(), passwordUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
