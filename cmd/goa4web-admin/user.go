package main

import (
	"flag"
	"fmt"
)

type userCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseUserCmd(parent *rootCmd, args []string) (*userCmd, error) {
	c := &userCmd{rootCmd: parent}
	fs := flag.NewFlagSet("user", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *userCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing user command")
	}
	switch c.args[0] {
	case "add":
		cmd, err := parseUserAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "add-admin":
		cmd, err := parseUserAddAdminCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add-admin: %w", err)
		}
		return cmd.Run()
	case "make-admin":
		cmd, err := parseUserMakeAdminCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("make-admin: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseUserUpdateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseUserListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown user command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *userCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s user <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  add\tadd a user")
	fmt.Fprintln(w, "  add-admin\tadd a user with administrator rights")
	fmt.Fprintln(w, "  make-admin\tgrant administrator rights to a user")
	fmt.Fprintln(w, "  list\tlist users")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s user add -username bob -password secret\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s user list\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
