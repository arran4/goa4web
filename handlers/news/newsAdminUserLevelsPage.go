package news

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminUserRolesPage(w http.ResponseWriter, r *http.Request) {
	type RoleInfo struct {
		PermID int32
		Name   string
	}
	type UserInfo struct {
		ID       int32
		Username sql.NullString
		Email    string
		Roles    []RoleInfo
	}
	type Data struct {
		Users []UserInfo
		Roles []*db.Role
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "News Roles"
	data := Data{}

	queries := cd.Queries()
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}

	users, err := queries.AdminListAllUsers(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("AdminListAllUsers Error: %s", err)
		handlers.RenderErrorPage(w, r, err)
		return
	}
	userMap := make(map[int32]*UserInfo)
	for _, u := range users {
		userMap[u.Idusers] = &UserInfo{ID: u.Idusers, Username: u.Username, Email: u.Email}
	}

	rows, err := queries.GetUserRoles(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("getUsersPermissions Error: %s", err)
		handlers.RenderErrorPage(w, r, err)
		return
	}
	for _, row := range rows {
		u, ok := userMap[row.UsersIdusers]
		if !ok {
			u = &UserInfo{ID: row.UsersIdusers, Username: row.Username, Email: row.Email}
			userMap[row.UsersIdusers] = u
		}
		u.Roles = append(u.Roles, RoleInfo{PermID: row.IduserRoles, Name: row.Role})
	}

	for _, u := range userMap {
		data.Users = append(data.Users, *u)
	}
	sort.Slice(data.Users, func(i, j int) bool {
		return data.Users[i].Username.String < data.Users[j].Username.String
	})

	handlers.TemplateHandler(w, r, "adminUserRolesPage.gohtml", data)
}
