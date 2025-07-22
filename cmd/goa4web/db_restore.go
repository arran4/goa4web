package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/dbdrivers"
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

func (c *dbRestoreCmd) Run() error {
	if c.File == "" {
		return fmt.Errorf("file required")
	}
	cfg := c.rootCmd.cfg
	c.rootCmd.Verbosef("restoring from %s", c.File)
	if err := dbdrivers.Restore(cfg.DBDriver, cfg.DBConn, c.File); err != nil {
		return err
	}
	c.rootCmd.Infof("database restored from %s", c.File)
	return nil
}
