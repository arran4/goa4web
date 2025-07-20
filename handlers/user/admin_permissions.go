package user

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"sort"
	"strconv"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

func adminUsersPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows  []*db.GetPermissionsWithUsersRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
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
}

// PermissionUserAllowTask grants a user permission.
type PermissionUserAllowTask struct{ tasks.TaskString }

var permissionUserAllowTask = &PermissionUserAllowTask{TaskString: TaskUserAllow}

func (PermissionUserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/permissions",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         role,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

// PermissionUserDisallowTask removes a user's permission.
type PermissionUserDisallowTask struct{ tasks.TaskString }

var permissionUserDisallowTask = &PermissionUserDisallowTask{TaskString: TaskUserDisallow}

func (PermissionUserDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/permissions",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

// PermissionUpdateTask updates an existing permission entry.
type PermissionUpdateTask struct{ tasks.TaskString }

var permissionUpdateTask = &PermissionUpdateTask{TaskString: TaskUpdate}

func (PermissionUpdateTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	role := r.PostFormValue("role")

	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/permissions",
	}

	if id, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.UpdatePermission(r.Context(), db.UpdatePermissionParams{
		IduserRoles: int32(id),
		Name:        role,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("UpdatePermission: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
