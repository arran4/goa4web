package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminUserListPage struct{}

func (p *AdminUserListPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminUserListPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Users"
	if _, err := cd.AdminListUsers(); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	AdminUserListPageTmpl.Handler(struct{}{}).ServeHTTP(w, r)
}

func (p *AdminUserListPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Users", "/admin/user", &AdminPage{}
}

func (p *AdminUserListPage) PageTitle() string {
	return "Users"
}

var _ common.Page = (*AdminUserListPage)(nil)
var _ tasks.Task = (*AdminUserListPage)(nil)
var _ http.Handler = (*AdminUserListPage)(nil)

const AdminUserListPageTmpl tasks.Template = "admin/userList.gohtml"
