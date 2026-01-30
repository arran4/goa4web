package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/dbops"
	"github.com/arran4/goa4web/internal/tasks"
)

// DBRestoreTask handles restoring database backups for administrators.
type DBRestoreTask struct{ tasks.TaskString }

var dbRestoreTask = &DBRestoreTask{TaskString: TaskDBRestore}

var _ tasks.Task = (*DBRestoreTask)(nil)
var _ tasks.AuditableTask = (*DBRestoreTask)(nil)

func (DBRestoreTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	if err := r.ParseMultipartForm(dbRestoreUploadMaxBytes); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if r.PostFormValue("confirm") != dbConfirmValue {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("confirmation required"))
	}
	upload, header, err := r.FormFile("backup")
	if err != nil {
		return fmt.Errorf("backup file required %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	defer upload.Close()
	tmpFile, err := os.CreateTemp("", "goa4web-restore-*.sql")
	if err != nil {
		return fmt.Errorf("create temp restore file %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	tmpPath := tmpFile.Name()
	if _, err := io.Copy(tmpFile, upload); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("copy restore file %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp restore file %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := dbops.RestoreDatabase(cd.DBRegistry(), cd.Config, tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("restore database %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	os.Remove(tmpPath)
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["Filename"] = header.Filename
	}
	data := struct {
		Messages []string
		Back     string
	}{
		Messages: []string{"Database restored successfully."},
		Back:     "/admin/db/restore",
	}
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

// AuditRecord summarizes a database restore.
func (DBRestoreTask) AuditRecord(data map[string]any) string {
	if name, ok := data["Filename"].(string); ok && name != "" {
		return fmt.Sprintf("database restored from %s", name)
	}
	return "database restored"
}
