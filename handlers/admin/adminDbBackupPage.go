package admin

import (
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminDBBackupPage struct{}

func (p *AdminDBBackupPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	AdminDBBackupPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminDBBackupPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Database Backup", "/admin/db/backup", &AdminPage{}
}

func (p *AdminDBBackupPage) PageTitle() string {
	return "Database Backup"
}

var _ common.Page = (*AdminDBBackupPage)(nil)
var _ http.Handler = (*AdminDBBackupPage)(nil)

// AdminDBBackupPageTmpl renders the admin database backup page.
const AdminDBBackupPageTmpl tasks.Template = "admin/dbBackupPage.gohtml"
