package writings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func UserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows  []*db.GetUserRolesRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
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

	handlers.TemplateHandler(w, r, "usersPermissionsPage.gohtml", data)
}

func UsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
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
		Back:     "/writings",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         role,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	} else if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["targetUserID"] = u.Idusers
			evt.Data["Username"] = u.Username.String
			evt.Data["Role"] = level
		}
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/writings",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else {
		id, username, role, err2 := roleInfoByPermID(r.Context(), queries, int32(permidi))
		if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
		} else if err2 == nil {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				if evt := cd.Event(); evt != nil {
					if evt.Data == nil {
						evt.Data = map[string]any{}
					}
					evt.Data["targetUserID"] = id
					evt.Data["Username"] = username
					evt.Data["Role"] = role
				}
			}
		} else {
			log.Printf("lookup role: %v", err2)
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func roleInfoByPermID(ctx context.Context, q *db.Queries, id int32) (int32, string, string, error) {
	rows, err := q.GetPermissionsWithUsers(ctx, db.GetPermissionsWithUsersParams{Username: sql.NullString{}})
	if err != nil {
		return 0, "", "", err
	}
	for _, row := range rows {
		if row.IduserRoles == id {
			return row.UsersIdusers, row.Username.String, row.Name, nil
		}
	}
	return 0, "", "", sql.ErrNoRows
}
