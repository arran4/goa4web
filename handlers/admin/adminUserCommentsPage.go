package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// adminUserCommentsPage lists all comments posted by a user.
func adminUserCommentsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cpu := cd.CurrentProfileUser()
	if cpu == nil || cpu.Idusers == 0 {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	if _, err := cd.AdminCommentsByUser(cpu.Idusers); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Comments by %s", cpu.Username.String)
	AdminUserCommentsPageTmpl.Handle(w, r, struct{}{})
}

const AdminUserCommentsPageTmpl handlers.Page = "admin/userCommentsPage.gohtml"
