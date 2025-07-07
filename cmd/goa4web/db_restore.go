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
	inFile, err := os.Open(c.File)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()
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
		cmd = exec.Command("mysql", args...)
	case "postgres":
		if cfg.DBConn == "" {
			return fmt.Errorf("connection string required")
		}
		cmd = exec.Command("psql", cfg.DBConn)
	case "sqlite3":
		path := cfg.DBConn
		if strings.HasPrefix(path, "file:") {
			if u, err := url.Parse(path); err == nil {
				path = u.Path
			}
		}
		cmd = exec.Command("sqlite3", path)
		cmd.Stdin = inFile
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("sqlite restore: %w", err)
		}
		if c.rootCmd.Verbosity > 0 {
			fmt.Printf("database restored from %s\n", c.File)
		}
		return nil
	default:
		return fmt.Errorf("restore not supported for driver %s", cfg.DBDriver)
	}
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("database restored from %s\n", c.File)
	}
	return nil
}
