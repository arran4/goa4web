package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminRoleEditFormPage struct{}

func (p *AdminRoleEditFormPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	role, err := cd.SelectedRole()
	if err != nil || role == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("role not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Role: %s", role.Name)

	id := cd.SelectedRoleID()
	groups, err := buildGrantGroups(r.Context(), cd, id)
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	data := struct {
		Role        *db.Role
		GrantGroups []GrantGroup
	}{Role: role, GrantGroups: groups}
	AdminRoleEditPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminRoleEditFormPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Edit Role", "", &AdminRolesPage{} // Or AdminRolePage if we can pass ID
}

func (p *AdminRoleEditFormPage) PageTitle() string {
	return "Edit Role"
}

var _ common.Page = (*AdminRoleEditFormPage)(nil)
var _ http.Handler = (*AdminRoleEditFormPage)(nil)

const AdminRoleEditPageTmpl tasks.Template = "admin/roleEditPage.gohtml"

// adminRoleEditSavePage persists role updates.
func adminRoleEditSavePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	id := cd.SelectedRoleID()
	if err := r.ParseForm(); err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	name := r.PostFormValue("name")
	canLogin := r.PostFormValue("can_login") != ""
	isAdmin := r.PostFormValue("is_admin") != ""
	privateLabels := r.PostFormValue("private_labels") != ""

	data := struct {
		Errors []string
		Back   string
	}{Back: fmt.Sprintf("/admin/role/%d", id)}

	role, err := queries.AdminGetRoleByID(r.Context(), id)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("get role: %w", err).Error())
		RunTaskPageTmpl.Handle(w, r, data)
		return
	}

	if err := queries.AdminUpdateRole(r.Context(), db.AdminUpdateRoleParams{
		Name:                   name,
		CanLogin:               canLogin,
		IsAdmin:                isAdmin,
		PrivateLabels:          privateLabels,
		PublicProfileAllowedAt: role.PublicProfileAllowedAt,
		ID:                     id,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update role: %w", err).Error())
	}
	RunTaskPageTmpl.Handle(w, r, data)
}
