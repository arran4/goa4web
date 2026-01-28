package admin

import (
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	if _, err := cd.AdminDashboardStats(); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("database not available"))
		return
	}
	AdminPageTmpl.Handle(w, r, struct{}{})
}

const AdminPageTmpl tasks.Template = "admin/page.gohtml"
