package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

type userPasswordGenerateResetCmd struct {
	*userCmd
	fs       *flag.FlagSet
	username string
	userID   int
}

func parseUserPasswordGenerateResetCmd(parent *userCmd, args []string) (*userPasswordGenerateResetCmd, error) {
	c := &userPasswordGenerateResetCmd{userCmd: parent}
	fs := flag.NewFlagSet("generate-reset", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.username, "username", "", "Username")
	fs.IntVar(&c.userID, "user-id", 0, "User ID")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userPasswordGenerateResetCmd) Run() error {
	ctx := c.rootCmd.Context()

	d, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}
	defer d.Close()

	queries := db.New(d)
	return generatePasswordReset(ctx, queries, c.rootCmd.cfg, c.userID, c.username)
}

// Usage prints command usage information with examples.
func (c *userPasswordGenerateResetCmd) Usage() {
	executeUsage(c.fs.Output(), "user_password_generate_reset_usage.txt", c)
}

func (c *userPasswordGenerateResetCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userPasswordGenerateResetCmd)(nil)
