package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/go-sql-driver/mysql"
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
	var cmd *exec.Cmd
	switch cfg.DBDriver {
	case "mysql":
		if cfg.DBConn == "" {
			return fmt.Errorf("connection string required")
		}
		mcfg, err := mysql.ParseDSN(cfg.DBConn)
		if err != nil {
			return fmt.Errorf("parse DSN: %w", err)
		}
		host, port, _ := strings.Cut(mcfg.Addr, ":")
		args := []string{
			"-h", host,
			"-P", port,
			"-u", mcfg.User,
			fmt.Sprintf("-p%s", mcfg.Passwd),
			mcfg.DBName,
		}
		cmd = exec.Command("mysqldump", args...)
	case "postgres":
		if cfg.DBConn == "" {
			return fmt.Errorf("connection string required")
		}
		cmd = exec.Command("pg_dump", "--dbname="+cfg.DBConn)
	case "sqlite3":
		path := cfg.DBConn
		if strings.HasPrefix(path, "file:") {
			if u, err := url.Parse(path); err == nil {
				path = u.Path
			}
		}
		cmd = exec.Command("sqlite3", path, ".dump")
	default:
		return fmt.Errorf("backup not supported for driver %s", cfg.DBDriver)
	}
	outFile, err := os.Create(c.File)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()
	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("database backup written to %s\n", c.File)
	}
	return nil
}
