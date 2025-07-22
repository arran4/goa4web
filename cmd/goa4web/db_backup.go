package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/dbdrivers"
)

// dbBackupCmd implements "db backup".
type dbBackupCmd struct {
	*dbCmd
	fs   *flag.FlagSet
	File string
	args []string
}

func parseDbBackupCmd(parent *dbCmd, args []string) (*dbBackupCmd, error) {
	c := &dbBackupCmd{dbCmd: parent}
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.StringVar(&c.File, "file", "", "output SQL file")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *dbBackupCmd) Run() error {
	if c.File == "" {
		return fmt.Errorf("file required")
	}
	cfg := c.rootCmd.cfg
	c.rootCmd.Verbosef("creating backup using %s", cfg.DBDriver)
	if err := dbdrivers.Backup(cfg.DBDriver, cfg.DBConn, c.File); err != nil {
		return err
	}
	c.rootCmd.Infof("database backup written to %s", c.File)
	return nil
}
