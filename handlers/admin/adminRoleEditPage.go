package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminRoleEditFormPage shows a form to update a role.
func adminRoleEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	role, err := cd.SelectedRole()
	if err != nil || role == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("role not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Role %s", role.Name)

	id := cd.SelectedRoleID()
	groups, err := buildGrantGroups(r.Context(), cd, id)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	data := struct {
		*common.CoreData
		Role        *db.Role
		GrantGroups []GrantGroup
	}{CoreData: cd, Role: role, GrantGroups: groups}
	handlers.TemplateHandler(w, r, "roleEditPage.gohtml", data)
}

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

	data := struct {
		*common.CoreData
		Errors []string
		Back   string
	}{CoreData: cd, Back: fmt.Sprintf("/admin/role/%d", id)}

	if err := queries.AdminUpdateRole(r.Context(), db.AdminUpdateRoleParams{
		Name:     name,
		CanLogin: canLogin,
		IsAdmin:  isAdmin,
		ID:       id,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update role: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
