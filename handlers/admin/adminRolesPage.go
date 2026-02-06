package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminRolesPage struct{}

func (p *AdminRolesPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminRolesPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Roles []*db.Role
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Roles"
	roles, err := cd.AllRoles()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	data := Data{Roles: roles}
	AdminRolesPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminRolesPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Roles", "/admin/roles", &AdminPage{}
}

func (p *AdminRolesPage) PageTitle() string {
	return "Admin Roles"
}

var _ common.Page = (*AdminRolesPage)(nil)
var _ tasks.Task = (*AdminRolesPage)(nil)
var _ http.Handler = (*AdminRolesPage)(nil)

const AdminRolesPageTmpl tasks.Template = "admin/rolesPage.gohtml"
