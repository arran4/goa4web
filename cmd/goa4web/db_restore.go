package main

import (
	"flag"
	"fmt"
)

// dbRestoreCmd implements "db restore".
type dbRestoreCmd struct {
	*dbCmd
	fs   *flag.FlagSet
	File string
}

func parseDbRestoreCmd(parent *dbCmd, args []string) (*dbRestoreCmd, error) {
	c := &dbRestoreCmd{dbCmd: parent}
	c.fs = newFlagSet("restore")
	c.fs.StringVar(&c.File, "file", "", "SQL file to restore")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *dbRestoreCmd) Usage() {
	executeUsage(c.fs.Output(), "db_restore_usage.txt", c)
}

func (c *dbRestoreCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*dbRestoreCmd)(nil)

func (c *dbRestoreCmd) Run() error {
	if c.File == "" {
		return fmt.Errorf("file required")
	}
	cfg := c.rootCmd.cfg
	c.rootCmd.Verbosef("restoring from %s", c.File)
	if err := c.rootCmd.dbReg.Restore(cfg.DBDriver, cfg.DBConn, c.File); err != nil {
		return err
	}
	c.rootCmd.Infof("database restored from %s", c.File)
	return nil
}
