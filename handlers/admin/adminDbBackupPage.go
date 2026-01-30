package admin

import (
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminDBBackupPage renders the database backup page.
func AdminDBBackupPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Backup"

	type Data struct {
		FileName string
		TaskName string
	}

	data := Data{
		FileName: defaultBackupFilename(time.Now()),
		TaskName: string(TaskDBBackup),
	}

	AdminDBBackupPageTmpl.Handle(w, r, data)
}

// AdminDBBackupPageTmpl renders the admin database backup page.
const AdminDBBackupPageTmpl tasks.Template = "admin/dbBackupPage.gohtml"
