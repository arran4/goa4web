package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminPageTask struct{}

func (t *AdminPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	if _, err := cd.AdminDashboardStats(); err != nil {
		return err
	}
	return AdminPageTmpl.Handler(struct{}{})
}

func (t *AdminPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Admin", "/admin", nil
}

// Ensure interface implementation
var _ tasks.Task = (*AdminPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminPageTask)(nil)

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

type AdminRolePageBreadcrumb struct {
	RoleName string
	RoleID   int32
}

func (p *AdminRolePageBreadcrumb) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Role %s", p.RoleName), "", &AdminRolesPageTask{}
}
