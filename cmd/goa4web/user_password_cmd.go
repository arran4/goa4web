package main

import (
	_ "embed"
	"flag"
	"fmt"
)

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
	if err := usageIfHelp(c.fs, c.args); err != nil {
		return err
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
	case "generate-reset":
		cmd, err := parseUserPasswordGenerateResetCmd(c.userCmd, c.args[1:])
		if err != nil {
			return fmt.Errorf("generate-reset: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown password command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *userPasswordCmd) Usage() {
	executeUsage(c.fs.Output(), "user_password_usage.txt", c)
}

func (c *userPasswordCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userPasswordCmd)(nil)
