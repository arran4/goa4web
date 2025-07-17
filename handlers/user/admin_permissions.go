package user

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func adminUsersPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Rows  []*db.GetPermissionsWithUsersRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecommon.CoreData),
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows, err := queries.GetPermissionsWithUsers(r.Context(), db.GetPermissionsWithUsersParams{Username: sql.NullString{}})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Username.String < rows[j].Username.String
	})
	data.Rows = rows

	common.TemplateHandler(w, r, "usersPermissionsPage.gohtml", data)
}

func adminUsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecommon.CoreData),
		Back:     "/admin/users/permissions",
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

func adminUsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecommon.CoreData),
		Back:     "/admin/users/permissions",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func adminUsersPermissionsUpdatePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	level := r.PostFormValue("role")

	data := struct {
		*corecommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecommon.CoreData),
		Back:     "/admin/users/permissions",
	}

	if id, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.UpdatePermission(r.Context(), db.UpdatePermissionParams{
		IduserRoles: int32(id),
		Name:        level,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("UpdatePermission: %w", err).Error())
	}

	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
