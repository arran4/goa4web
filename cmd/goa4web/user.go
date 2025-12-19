package main

import (
	"flag"
	"fmt"
)

type userCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseUserCmd(parent *rootCmd, args []string) (*userCmd, error) {
	c := &userCmd{rootCmd: parent}
	c.fs = newFlagSet("user")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing user command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "add":
		cmd, err := parseUserAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "add-admin":
		cmd, err := parseUserAddAdminCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add-admin: %w", err)
		}
		return cmd.Run()
	case "make-admin":
		cmd, err := parseUserMakeAdminCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("make-admin: %w", err)
		}
		return cmd.Run()
	case "add-role":
		cmd, err := parseUserAddRoleCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add-role: %w", err)
		}
		return cmd.Run()
	case "remove-role":
		cmd, err := parseUserRemoveRoleCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("remove-role: %w", err)
		}
		return cmd.Run()
	case "list-roles":
		cmd, err := parseUserListRolesCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list-roles: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseUserListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "deactivate":
		cmd, err := parseUserDeactivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("deactivate: %w", err)
		}
		return cmd.Run()
	case "activate":
		cmd, err := parseUserActivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("activate: %w", err)
		}
		return cmd.Run()
	case "approve":
		cmd, err := parseUserApproveCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("approve: %w", err)
		}
		return cmd.Run()
	case "reject":
		cmd, err := parseUserRejectCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("reject: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseUserCommentsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	case "roles":
		cmd, err := parseUserRolesCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("roles: %w", err)
		}
		return cmd.Run()
	case "password":
		cmd, err := parseUserPasswordCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("password: %w", err)
		}
		return cmd.Run()
	case "profile":
		cmd, err := parseUserProfileCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("profile: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown user command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *userCmd) Usage() {
	executeUsage(c.fs.Output(), "user_usage.txt", c)
}

func (c *userCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userCmd)(nil)
