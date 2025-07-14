package main

import (
	"flag"
	"fmt"
)

// dbCmd handles database utilities like migrations.
type dbCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseDbCmd(parent *rootCmd, args []string) (*dbCmd, error) {
	fs := flag.NewFlagSet("db", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &dbCmd{rootCmd: parent, fs: fs, args: fs.Args()}, nil
}

func (c *dbCmd) Run() error {
	if len(c.args) == 0 {
		return fmt.Errorf("missing db command")
	}
	switch c.args[0] {
	case "migrate":
		cmd, err := parseDbMigrateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
		return cmd.Run()
	case "backup":
		cmd, err := parseDbBackupCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("backup: %w", err)
		}
		return cmd.Run()
	case "restore":
		cmd, err := parseDbRestoreCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("restore: %w", err)
		}
		return cmd.Run()
	case "seed":
		cmd, err := parseDbSeedCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("seed: %w", err)
		}
		return cmd.Run()
	default:
		return fmt.Errorf("unknown db command %q", c.args[0])
	}
}
