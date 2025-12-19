package main

import (
	"flag"
	"fmt"
)

// dbCmd handles database utilities like migrations.
type dbCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseDbCmd(parent *rootCmd, args []string) (*dbCmd, error) {
	c := &dbCmd{rootCmd: parent}
	c.fs = newFlagSet("db")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *dbCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing db command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "create":
		cmd, err := parseDbCreateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return cmd.Run()
	case "seed":
		cmd, err := parseDbSeedCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("seed: %w", err)
		}
		return cmd.Run()
	case "setup":
		cmd, err := parseDbSetupCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		return cmd.Run()
	case "migrate":
		cmd, err := parseDbMigrateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
		return cmd.Run()
	case "backup":
		cmd, err := parseDbBackupCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("backup: %w", err)
		}
		return cmd.Run()
	case "restore":
		cmd, err := parseDbRestoreCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("restore: %w", err)
		}
		return cmd.Run()
	case "show":
		cmd, err := parseDbShowCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("show: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown db command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *dbCmd) Usage() {
	executeUsage(c.fs.Output(), "db_usage.txt", c)
}

func (c *dbCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*dbCmd)(nil)
