package main

import (
	"flag"
	"fmt"

	dbdrivers "github.com/arran4/goa4web/internal/dbdrivers"
)

// dbRestoreCmd implements "db restore".
type dbRestoreCmd struct {
	*dbCmd
	fs   *flag.FlagSet
	File string
	args []string
}

func parseDbRestoreCmd(parent *dbCmd, args []string) (*dbRestoreCmd, error) {
	c := &dbRestoreCmd{dbCmd: parent}
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.StringVar(&c.File, "file", "", "SQL file to restore")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *dbRestoreCmd) Run() error {
	if c.File == "" {
		return fmt.Errorf("file required")
	}
	cfg := c.rootCmd.cfg
	if err := dbdrivers.Restore(cfg.DBDriver, cfg.DBConn, c.File); err != nil {
		return err
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("database restored from %s\n", c.File)
	}
	return nil
}
