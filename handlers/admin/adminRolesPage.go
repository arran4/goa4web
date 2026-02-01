package admin

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminRolesPageTask struct{}

func (t *AdminRolesPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		Roles []*db.Role
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin Roles"
	roles, err := cd.AllRoles()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	data := Data{Roles: roles}
	return AdminRolesPageTmpl.Handler(data)
}

func (t *AdminRolesPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Roles", "/admin/roles", &AdminPageTask{}
}

var _ tasks.Task = (*AdminRolesPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminRolesPageTask)(nil)

const AdminRolesPageTmpl tasks.Template = "admin/rolesPage.gohtml"
