package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	args := []string{
		"-h", cfg.DBHost,
		"-P", cfg.DBPort,
		"-u", cfg.DBUser,
		fmt.Sprintf("-p%s", cfg.DBPass),
		cfg.DBName,
	}
	cmd := exec.Command("mysqldump", args...)
	outFile, err := os.Create(c.File)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()
	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("database backup written to %s\n", c.File)
	}
	return nil
}
