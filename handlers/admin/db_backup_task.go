package admin

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/dbops"
	"github.com/arran4/goa4web/internal/tasks"
)

// DBBackupTask handles creating database backups for administrators.
type DBBackupTask struct{ tasks.TaskString }

var dbBackupTask = &DBBackupTask{TaskString: TaskDBBackup}

var _ tasks.Task = (*DBBackupTask)(nil)
var _ tasks.AuditableTask = (*DBBackupTask)(nil)

func (DBBackupTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if r.PostFormValue("confirm") != dbConfirmValue {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("confirmation required"))
	}
	filename := safeBackupFilename(r.PostFormValue("filename"), defaultBackupFilename(time.Now()))
	tmpFile, err := os.CreateTemp("", "goa4web-backup-*.sql")
	if err != nil {
		return fmt.Errorf("create temp backup file %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp backup file %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := dbops.BackupDatabase(cd.DBRegistry(), cd.Config, tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("backup database %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["Filename"] = filename
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer os.Remove(tmpPath)
		file, err := os.Open(tmpPath)
		if err != nil {
			handlers.RenderErrorPage(w, r, fmt.Errorf("open backup file: %w", err))
			return
		}
		defer file.Close()
		modTime := time.Now()
		if info, err := file.Stat(); err == nil {
			modTime = info.ModTime()
		}
		w.Header().Set("Content-Type", "application/sql")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		http.ServeContent(w, r, filename, modTime, file)
	})
}

// AuditRecord summarizes a database backup.
func (DBBackupTask) AuditRecord(data map[string]any) string {
	if name, ok := data["Filename"].(string); ok && name != "" {
		return fmt.Sprintf("database backup downloaded as %s", name)
	}
	return "database backup downloaded"
}
