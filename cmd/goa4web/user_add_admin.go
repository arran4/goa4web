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
}

func parseUserAddAdminCmd(parent *userCmd, args []string) (*userAddAdminCmd, error) {
	c := &userAddAdminCmd{userCmd: parent}
	fs, _, err := parseFlags("add-admin", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
		fs.StringVar(&c.Email, "email", "", "email address")
		fs.StringVar(&c.Password, "password", "", "password (leave empty to prompt)")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *userAddAdminCmd) Usage() {
	executeUsage(c.fs.Output(), "user_add_admin_usage.txt", c)
}

func (c *userAddAdminCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userAddAdminCmd)(nil)

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
