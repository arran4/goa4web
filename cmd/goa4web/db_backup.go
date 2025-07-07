package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers"
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
	h := dbhandlers.BackupFor(cfg.DBDriver)
	if h == nil {
		return fmt.Errorf("backup not supported for driver %s", cfg.DBDriver)
	}
	if err := h.Backup(cfg, c.File); err != nil {
		return err
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("database backup written to %s\n", c.File)
	}
	return nil
}
