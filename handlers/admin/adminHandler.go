package admin

import (
	"fmt"
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

const AdminPageTmpl handlers.Page = "admin/page.gohtml"
