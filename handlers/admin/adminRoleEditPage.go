package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminRoleEditFormPage shows a form to update a role.
func adminRoleEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	role, err := queries.AdminGetRoleByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}
	cd.PageTitle = fmt.Sprintf("Edit Role %s", role.Name)
	data := struct {
		*common.CoreData
		Role *db.Role
	}{CoreData: cd, Role: role}
	handlers.TemplateHandler(w, r, "roleEditPage.gohtml", data)
}

// adminRoleEditSavePage persists role updates.
func adminRoleEditSavePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
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

	if _, err := queries.DB().ExecContext(r.Context(), "UPDATE roles SET name=?, can_login=?, is_admin=? WHERE id=?", name, canLogin, isAdmin, id); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update role: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
