package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminPage struct{}

func (p *AdminPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	if _, err := cd.AdminDashboardStats(); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	AdminPageTmpl.Handler(struct{}{}).ServeHTTP(w, r)
}

func (p *AdminPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Admin", "/admin", nil
}

func (p *AdminPage) PageTitle() string {
	return "Admin"
}

// Ensure interface implementation
var _ tasks.Page = (*AdminPage)(nil)
var _ tasks.Task = (*AdminPage)(nil)
var _ http.Handler = (*AdminPage)(nil)

const AdminPageTmpl tasks.Template = "admin/page.gohtml"
