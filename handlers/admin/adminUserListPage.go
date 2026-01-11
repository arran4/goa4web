package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func adminUserListPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Users"
	if _, err := cd.AdminListUsers(); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	AdminUserListPageTmpl.Handle(w, r, struct{}{})
}

const AdminUserListPageTmpl handlers.Page = "admin/userList.gohtml"
