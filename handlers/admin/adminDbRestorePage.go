package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminDBRestorePage renders the database restore page.
func AdminDBRestorePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Restore"

	type Data struct {
		TaskName string
	}

	data := Data{
		TaskName: string(TaskDBRestore),
	}

	AdminDBRestorePageTmpl.Handle(w, r, data)
}

// AdminDBRestorePageTmpl renders the admin database restore page.
const AdminDBRestorePageTmpl tasks.Template = "admin/dbRestorePage.gohtml"
