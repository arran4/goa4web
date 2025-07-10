package mysql

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers"
	"github.com/arran4/goa4web/config"
)

type handler struct{}

func (handler) Backup(cfg config.RuntimeConfig, file string) error {
	if cfg.DBConn == "" {
		return fmt.Errorf("connection string required")
	}
	mcfg, err := mysql.ParseDSN(cfg.DBConn)
	if err != nil {
		return fmt.Errorf("parse DSN: %w", err)
	}
	host, port, _ := strings.Cut(mcfg.Addr, ":")
	args := []string{"-h", host, "-P", port, "-u", mcfg.User, fmt.Sprintf("-p%s", mcfg.Passwd), mcfg.DBName}
	cmd := exec.Command("mysqldump", args...)
	outFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()
	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	return nil
}

func (handler) Restore(cfg config.RuntimeConfig, file string) error {
	if cfg.DBConn == "" {
		return fmt.Errorf("connection string required")
	}
	inFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()
	mcfg, err := mysql.ParseDSN(cfg.DBConn)
	if err != nil {
		return fmt.Errorf("parse DSN: %w", err)
	}
	host, port, _ := strings.Cut(mcfg.Addr, ":")
	args := []string{"-h", host, "-P", port, "-u", mcfg.User, fmt.Sprintf("-p%s", mcfg.Passwd), mcfg.DBName}
	cmd := exec.Command("mysql", args...)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	return nil
}

// Register registers the mysql handlers with the default registry.
func Register() {
	dbhandlers.RegisterBackup("mysql", handler{})
	dbhandlers.RegisterRestore("mysql", handler{})
}
