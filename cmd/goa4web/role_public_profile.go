package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// rolePublicProfileCmd implements the "role public-profile" subcommand.
type rolePublicProfileCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	args []string
}

func parseRolePublicProfileCmd(parent *roleCmd, args []string) (*rolePublicProfileCmd, error) {
	c := &rolePublicProfileCmd{roleCmd: parent}
	fs := flag.NewFlagSet("public-profile", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *rolePublicProfileCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing public-profile subcommand")
	}
	if err := usageIfHelp(c.fs, c.args); err != nil {
		return err
	}
	switch c.args[0] {
	case "set":
		cmd, err := parseRolePublicProfileSetCmd(c, c.args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown public-profile subcommand %q", c.args[0])
	}
}

func (c *rolePublicProfileCmd) Usage() {
	executeUsage(c.fs.Output(), "role_public_profile_usage.txt", c)
}

func (c *rolePublicProfileCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*rolePublicProfileCmd)(nil)

type boolFlag struct {
	value bool
	set   bool
}

func (b *boolFlag) String() string {
	if !b.set {
		return ""
	}
	return strconv.FormatBool(b.value)
}

func (b *boolFlag) Set(value string) error {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	b.value = parsed
	b.set = true
	return nil
}

func (b *boolFlag) IsBoolFlag() bool { return true }

// rolePublicProfileSetCmd implements "role public-profile set".
type rolePublicProfileSetCmd struct {
	*rolePublicProfileCmd
	fs      *flag.FlagSet
	role    string
	enabled boolFlag
}

func parseRolePublicProfileSetCmd(parent *rolePublicProfileCmd, args []string) (*rolePublicProfileSetCmd, error) {
	c := &rolePublicProfileSetCmd{rolePublicProfileCmd: parent}
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.StringVar(&c.role, "role", "", "The role name to update.")
	fs.Var(&c.enabled, "enabled", "Whether to enable public profiles for the role.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.role == "" {
		return nil, fmt.Errorf("role name is required")
	}
	if !c.enabled.set {
		return nil, fmt.Errorf("enabled flag is required")
	}
	return c, nil
}

func (c *rolePublicProfileSetCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	q := db.New(sdb)
	ctx := c.rootCmd.ctx

	role, err := q.GetRoleByName(ctx, c.role)
	if err != nil {
		return fmt.Errorf("failed to get role %q: %w", c.role, err)
	}

	var ts sql.NullTime
	if c.enabled.value {
		ts = sql.NullTime{Time: time.Now(), Valid: true}
	}
	if err := q.AdminUpdateRolePublicProfileAllowed(ctx, db.AdminUpdateRolePublicProfileAllowedParams{
		PublicProfileAllowedAt: ts,
		ID:                     role.ID,
	}); err != nil {
		return fmt.Errorf("update role public profile: %w", err)
	}

	status := "disabled"
	if c.enabled.value {
		status = "enabled"
	}
	fmt.Printf("Role %q public profile access %s.\n", role.Name, status)
	fmt.Printf("Status: %s\n", status)
	return nil
}

func (c *rolePublicProfileSetCmd) Usage() {
	executeUsage(c.fs.Output(), "role_public_profile_set_usage.txt", c)
}

func (c *rolePublicProfileSetCmd) FlagGroups() []flagGroup {
	return append(c.roleCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*rolePublicProfileSetCmd)(nil)
