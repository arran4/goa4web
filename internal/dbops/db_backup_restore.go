package dbops

import (
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/dbdrivers"
)

// BackupDatabase validates inputs and writes a database backup to file.
func BackupDatabase(reg *dbdrivers.Registry, cfg *config.RuntimeConfig, file string) error {
	if reg == nil {
		return fmt.Errorf("database registry required")
	}
	if cfg == nil {
		return fmt.Errorf("runtime config required")
	}
	if file == "" {
		return fmt.Errorf("file required")
	}
	if cfg.DBDriver == "" {
		return fmt.Errorf("db driver required")
	}
	if cfg.DBConn == "" {
		return fmt.Errorf("db connection required")
	}
	return reg.Backup(cfg.DBDriver, cfg.DBConn, file)
}

// RestoreDatabase validates inputs and restores a database backup from file.
func RestoreDatabase(reg *dbdrivers.Registry, cfg *config.RuntimeConfig, file string) error {
	if reg == nil {
		return fmt.Errorf("database registry required")
	}
	if cfg == nil {
		return fmt.Errorf("runtime config required")
	}
	if file == "" {
		return fmt.Errorf("file required")
	}
	if cfg.DBDriver == "" {
		return fmt.Errorf("db driver required")
	}
	if cfg.DBConn == "" {
		return fmt.Errorf("db connection required")
	}
	return reg.Restore(cfg.DBDriver, cfg.DBConn, file)
}
