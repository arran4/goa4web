package blogs

import (
	"database/sql"
	"errors"
	"fmt"

	db "github.com/arran4/goa4web/internal/db"

	"net/http"
	"strconv"
	"strings"

	common "github.com/arran4/goa4web/handlers/common"
)

func GetPermissionsByUserIdAndSectionBlogsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(corecommon.KeyCoreData).(*CoreData)
	if !(cd.HasRole("content writer") || cd.HasRole("administrator")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	type Data struct {
		*CoreData
		Rows   []*db.GetUserRolesRow
		Filter string
		Roles  []*db.Role
	}

	data := Data{
		CoreData: cd,
		Filter:   r.URL.Query().Get("level"),
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
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

	if data.Filter != "" {
		filtered := rows[:0]
		for _, row := range rows {
			if row.Role == data.Filter {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	data.Rows = rows

	common.TemplateHandler(w, r, "userPermissionsPage.gohtml", data)
}

func UsersPermissionsPermissionUserAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
		Back:     "/blogs/bloggers",
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
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
		Back:     "/blogs/bloggers",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
	}
	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsBulkAllowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	names := strings.FieldsFunc(r.PostFormValue("usernames"), func(r rune) bool { return r == ',' || r == '\n' || r == ' ' || r == '\t' })
	level := r.PostFormValue("role")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
		Back:     "/blogs/bloggers",
	}

	for _, n := range names {
		if n == "" {
			continue
		}
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: n})
		if err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername %s: %w", n, err).Error())
			continue
		}
		if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
			UsersIdusers: u.Idusers,
			Name:         level,
		}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow %s: %w", n, err).Error())
		}
	}

	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func UsersPermissionsBulkDisallowPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	permids := r.PostForm["permid"]
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
		Back:     "/blogs/bloggers",
	}

	for _, id := range permids {
		if id == "" {
			continue
		}
		permidi, err := strconv.Atoi(id)
		if err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi %s: %w", id, err).Error())
			continue
		}
		if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("permissionUserDisallow %s: %w", id, err).Error())
		}
	}

	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
