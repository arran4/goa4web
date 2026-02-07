package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminUserCommentsPage struct{}

func (p *AdminUserCommentsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cpu := cd.CurrentProfileUser()
	if cpu == nil || cpu.Idusers == 0 {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	if _, err := cd.AdminCommentsByUser(cpu.Idusers); err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Comments by %s", cpu.Username.String)
	AdminUserCommentsPageTmpl.Handler(struct{}{}).ServeHTTP(w, r)
}

func (p *AdminUserCommentsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Comments", "", &AdminUserProfilePage{}
}

func (p *AdminUserCommentsPage) PageTitle() string {
	return "User Comments"
}

var _ common.Page = (*AdminUserCommentsPage)(nil)
var _ http.Handler = (*AdminUserCommentsPage)(nil)

const AdminUserCommentsPageTmpl tasks.Template = "admin/userCommentsPage.gohtml"
