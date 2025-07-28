package linker

import (
	"context"
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminUserRolesPage(w http.ResponseWriter, r *http.Request) {
	type PermissionUser struct {
		*db.GetUserRolesRow
		Username sql.NullString
		Email    sql.NullString
	}

	type Data struct {
		*common.CoreData
		UserLevels []*PermissionUser
		Search     string
		Roles      []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Search:   r.URL.Query().Get("search"),
	}
	data.CoreData.PageTitle = "User Roles"

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}
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

	handlers.TemplateHandler(w, r, "adminUserRolesPage.gohtml", data)
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
