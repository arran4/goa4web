package main

import (
	"flag"

	"fmt"
)

// userAddAdminCmd implements "user add-admin".
type userAddAdminCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
	Email    string
	Password string
	args     []string
}

func parseUserAddAdminCmd(parent *userCmd, args []string) (*userAddAdminCmd, error) {
	c := &userAddAdminCmd{userCmd: parent}
	fs := flag.NewFlagSet("add-admin", flag.ContinueOnError)
	fs.StringVar(&c.Username, "username", "", "username")
	fs.StringVar(&c.Email, "email", "", "email address")
	fs.StringVar(&c.Password, "password", "", "password (leave empty to prompt)")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *userAddAdminCmd) Run() error {
	pw := c.Password
	if pw == "" {
		var err error
		if pw, err = promptPassword(); err != nil {
			return fmt.Errorf("prompt password: %w", err)
		}
	}
	return createUser(c.userCmd.rootCmd, c.Username, c.Email, pw, true)
}
