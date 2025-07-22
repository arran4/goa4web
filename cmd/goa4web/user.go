package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/user_usage.txt
var userUsageTemplate string

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
	case "add-role":
		cmd, err := parseUserAddRoleCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add-role: %w", err)
		}
		return cmd.Run()
	case "remove-role":
		cmd, err := parseUserRemoveRoleCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("remove-role: %w", err)
		}
		return cmd.Run()
	case "list-roles":
		cmd, err := parseUserListRolesCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list-roles: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseUserListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "deactivate":
		cmd, err := parseUserDeactivateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("deactivate: %w", err)
		}
		return cmd.Run()
	case "activate":
		cmd, err := parseUserActivateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("activate: %w", err)
		}
		return cmd.Run()
	case "approve":
		cmd, err := parseUserApproveCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("approve: %w", err)
		}
		return cmd.Run()
	case "reject":
		cmd, err := parseUserRejectCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("reject: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseUserCommentsCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	case "roles":
		cmd, err := parseUserRolesCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("roles: %w", err)
		}
		return cmd.Run()
	case "password":
		cmd, err := parseUserPasswordCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("password: %w", err)
		}
		return cmd.Run()
	case "profile":
		cmd, err := parseUserProfileCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("profile: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown user command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *userCmd) Usage() {
	executeUsage(c.fs.Output(), userUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
