package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminDBRestorePage struct{}

func (p *AdminDBRestorePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Restore"

	type Data struct {
		TaskName string
	}

	data := Data{
		TaskName: string(TaskDBRestore),
	}

	AdminDBRestorePageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminDBRestorePage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Database Restore", "/admin/db/restore", &AdminPage{}
}

func (p *AdminDBRestorePage) PageTitle() string {
	return "Database Restore"
}

var _ common.Page = (*AdminDBRestorePage)(nil)
var _ http.Handler = (*AdminDBRestorePage)(nil)

// AdminDBRestorePageTmpl renders the admin database restore page.
const AdminDBRestorePageTmpl tasks.Template = "admin/dbRestorePage.gohtml"
