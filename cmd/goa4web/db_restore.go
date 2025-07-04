package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	args := []string{
		"-h", cfg.DBHost,
		"-P", cfg.DBPort,
		"-u", cfg.DBUser,
		fmt.Sprintf("-p%s", cfg.DBPass),
		cfg.DBName,
	}
	inFile, err := os.Open(c.File)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()
	cmd := exec.Command("mysql", args...)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("database restored from %s\n", c.File)
	}
	return nil
}
