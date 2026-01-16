package main

import (
	"fmt"
	"os"
)

type userEmailCmd struct {
	*userCmd
}

func parseUserEmailCmd(parent *userCmd, args []string) (*userEmailCmd, error) {
	c := &userEmailCmd{userCmd: parent}
	c.fs = newFlagSet("email")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userEmailCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.Usage()
		return fmt.Errorf("missing email subcommand")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}

	// This is a convenience command that redirects to the `email user` command.
	emailCmdArgs := []string{"user"}
	emailCmdArgs = append(emailCmdArgs, args...)

	emailCmd, err := parseEmailCmd(c.rootCmd, emailCmdArgs)
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}
	return emailCmd.Run()
}

func (c *userEmailCmd) Usage() {
	fmt.Fprintf(c.fs.Output(), "Usage: %s user email [subcommand] [flags]\n\nThis is a convenience command that redirects to `%s email user`.\n", os.Args[0], os.Args[0])
	c.fs.PrintDefaults()
}
