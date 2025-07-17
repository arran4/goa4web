package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func UserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecorecommon.CoreData
		Rows  []*db.GetUserRolesRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecorecommon.CoreData),
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows, err := queries.GetUserRoles(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	common.TemplateHandler(w, r, "usersPermissionsPage.gohtml", data)
}

func UsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	data := struct {
		*corecorecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecorecommon.CoreData),
		Back:     "/writings",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         level,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	}

	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*corecorecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecorecommon.CoreData),
		Back:     "/writings",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
