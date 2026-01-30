package admin

import (
	"path/filepath"
	"strings"
	"time"
)

const (
	// dbConfirmValue is the expected confirmation form value.
	dbConfirmValue = "yes"
	// dbRestoreUploadMaxBytes caps uploaded restore file sizes.
	dbRestoreUploadMaxBytes int64 = 100 << 20
)

func defaultBackupFilename(t time.Time) string {
	return "goa4web-backup-" + t.Format("20060102-150405") + ".sql"
}

func safeBackupFilename(name, fallback string) string {
	cleaned := strings.TrimSpace(name)
	if cleaned == "" {
		return fallback
	}
	base := filepath.Base(cleaned)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return fallback
	}
	return base
}
