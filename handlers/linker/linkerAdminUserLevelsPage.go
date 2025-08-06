package linker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

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
		Users  []UserInfo
		Search string
		Roles  []*db.Role
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		Search: r.URL.Query().Get("search"),
	}
	cd.PageTitle = "User Roles"

	queries := cd.Queries()
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}

	users, err := queries.AdminListAllUsers(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("AdminListAllUsers Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	userMap := make(map[int32]*UserInfo)
	for _, u := range users {
		userMap[u.Idusers] = &UserInfo{ID: u.Idusers, Username: u.Username, Email: u.Email}
	}

	rows, err := queries.GetUserRoles(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("getUsersPermissions Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
		if data.Search != "" && !strings.Contains(strings.ToLower(u.Username.String), strings.ToLower(data.Search)) {
			continue
		}
		data.Users = append(data.Users, *u)
	}
	sort.Slice(data.Users, func(i, j int) bool {
		return data.Users[i].Username.String < data.Users[j].Username.String
	})

	handlers.TemplateHandler(w, r, "adminUserRolesPage.gohtml", data)
}

func roleInfoByPermID(ctx context.Context, q db.Querier, id int32) (int32, string, string, error) {
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
