package linker

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/templates"
)

func AdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type PermissionUser struct {
		*db.GetUserRolesRow
		Username sql.NullString
		Email    sql.NullString
	}

	type Data struct {
		*corecommon.CoreData
		UserLevels []*PermissionUser
		Search     string
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	rows, err := queries.GetUserRoles(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getUsersPermissions Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	var perms []*PermissionUser
	for _, p := range rows {
		row, err := queries.GetUserById(r.Context(), p.UsersIdusers)
		if err != nil {
			log.Printf("GetUserById Error: %s", err)
			continue
		}
		perms = append(perms, &PermissionUser{GetUserRolesRow: p, Username: row.Username, Email: row.Email})
	}

	if data.Search != "" {
		q := strings.ToLower(data.Search)
		var filtered []*PermissionUser
		for _, row := range perms {
			if strings.Contains(strings.ToLower(row.Username.String), q) {
				filtered = append(filtered, row)
			}
		}
		perms = filtered
	}
	data.UserLevels = perms

	CustomLinkerIndex(data.CoreData, r)
	if err := templates.RenderTemplate(w, "adminUserLevelsPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminUserLevelsAllowActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	usernames := r.PostFormValue("usernames")
	level := r.PostFormValue("role")
	fields := strings.FieldsFunc(usernames, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	for _, username := range fields {
		if username == "" {
			continue
		}
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("GetUserByUsername Error: %s", err)
			continue
		}
		if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
			UsersIdusers: u.Idusers,
			Name:         level,
		}); err != nil {
			log.Printf("permissionUserAllow Error: %s", err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminUserLevelsRemoveActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	r.ParseForm()
	ids := r.Form["permids"]
	if len(ids) == 0 {
		if id := r.PostFormValue("permid"); id != "" {
			ids = append(ids, id)
		}
	}
	for _, idStr := range ids {
		permid, _ := strconv.Atoi(idStr)
		if err := queries.DeleteUserRole(r.Context(), int32(permid)); err != nil {
			log.Printf("permissionUserDisallow Error: %s", err)
		}
	}
	common.TaskDoneAutoRefreshPage(w, r)
}
