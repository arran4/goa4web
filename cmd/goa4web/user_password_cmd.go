package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/user_password_usage.txt
var userPasswordUsageTemplate string

// userPasswordCmd handles reset password requests.
type userPasswordCmd struct {
	*userCmd
	fs   *flag.FlagSet
	args []string
}

func parseUserPasswordCmd(parent *userCmd, args []string) (*userPasswordCmd, error) {
	c := &userPasswordCmd{userCmd: parent}
	fs := flag.NewFlagSet("password", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *userPasswordCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing password command")
	}
	switch c.args[0] {
	case "clear-expired":
		cmd, err := parseUserPasswordClearExpiredCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("clear-expired: %w", err)
		}
		return cmd.Run()
	case "clear-user":
		cmd, err := parseUserPasswordClearUserCmd(c, c.args[1:])
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
func (c *userPasswordCmd) Usage() {
	executeUsage(c.fs.Output(), userPasswordUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
