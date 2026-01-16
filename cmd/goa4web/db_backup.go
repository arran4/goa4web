package main

import (
	"flag"
	"fmt"
)

// dbBackupCmd implements "db backup".
type dbBackupCmd struct {
	*dbCmd
	fs   *flag.FlagSet
	File string
}

func parseDbBackupCmd(parent *dbCmd, args []string) (*dbBackupCmd, error) {
	c := &dbBackupCmd{dbCmd: parent}
	c.fs = newFlagSet("backup")
	c.fs.StringVar(&c.File, "file", "", "output SQL file")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *dbBackupCmd) Usage() {
	executeUsage(c.fs.Output(), "db_backup_usage.txt", c)
}

func (c *dbBackupCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*dbBackupCmd)(nil)

func (c *dbBackupCmd) Run() error {
	if c.File == "" {
		return fmt.Errorf("file required")
	}
	cfg := c.rootCmd.cfg
	c.rootCmd.Verbosef("creating backup using %s", cfg.DBDriver)
	if err := c.rootCmd.dbReg.Backup(cfg.DBDriver, cfg.DBConn, c.File); err != nil {
		return err
	}
	c.rootCmd.Infof("database backup written to %s", c.File)
	return nil
}
