package sqlite

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/arran4/goa4web/cmd/goa4web/dbhandlers"
	"github.com/arran4/goa4web/config"
)

type handler struct{}

func pathFromConn(conn string) string {
	path := conn
	if strings.HasPrefix(path, "file:") {
		if u, err := url.Parse(path); err == nil {
			path = u.Path
		}
	}
	return path
}

func (handler) Backup(cfg config.RuntimeConfig, file string) error {
	path := pathFromConn(cfg.DBConn)
	cmd := exec.Command("sqlite3", path, ".dump")
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
	path := pathFromConn(cfg.DBConn)
	inFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()
	cmd := exec.Command("sqlite3", path)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sqlite restore: %w", err)
	}
	return nil
}

// Register registers the sqlite handlers with the default registry.
func Register() {
	dbhandlers.RegisterBackup("sqlite3", handler{})
	dbhandlers.RegisterRestore("sqlite3", handler{})
}
