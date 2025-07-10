package postgres

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers"
	"github.com/arran4/goa4web/config"
)

type handler struct{}

func (handler) Backup(cfg config.RuntimeConfig, file string) error {
	if cfg.DBConn == "" {
		return fmt.Errorf("connection string required")
	}
	cmd := exec.Command("pg_dump", "--dbname="+cfg.DBConn)
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
	cmd := exec.Command("psql", cfg.DBConn)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	return nil
}

// Register registers the postgres handlers with the default registry.
func Register() {
	dbhandlers.RegisterBackup("postgres", handler{})
	dbhandlers.RegisterRestore("postgres", handler{})
}
