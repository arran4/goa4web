package news

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func NewsUserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData)
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	type Data struct {
		*handlers.CoreData
		Rows  []*db.GetUserRolesRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: cd,
	}

	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	if roles, err := cd.AllRoles(); err == nil {
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

	handlers.TemplateHandler(w, r, "userPermissionsPage.gohtml", data)
}

func NewsUsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	data := struct {
		*handlers.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Back:     "/news",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         level,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func NewsUsersPermissionsDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*handlers.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Back:     "/news",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
