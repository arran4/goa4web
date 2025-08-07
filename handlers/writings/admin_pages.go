package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// AdminWritingsPage renders the writings admin index with role summaries.
func AdminWritingsPage(w http.ResponseWriter, r *http.Request) {
	type RoleInfo struct {
		ID       int32
		Username sql.NullString
		Email    string
		Roles    []string
	}
	type Data struct {
		CanPost   bool
		UserRoles []RoleInfo
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writings Admin"
	data := Data{CanPost: cd.HasGrant("writing", "post", "edit", 0) && cd.AdminMode}

	queries := cd.Queries()
	rows, err := queries.GetUserRoles(r.Context())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	userMap := make(map[int32]*RoleInfo)
	for _, row := range rows {
		if row.Role != "administrator" && row.Role != "content writer" {
			continue
		}
		u, ok := userMap[row.UsersIdusers]
		if !ok {
			u = &RoleInfo{ID: row.UsersIdusers, Username: row.Username, Email: row.Email}
			userMap[row.UsersIdusers] = u
		}
		u.Roles = append(u.Roles, row.Role)
	}
	for _, u := range userMap {
		data.UserRoles = append(data.UserRoles, *u)
	}
	sort.Slice(data.UserRoles, func(i, j int) bool {
		return data.UserRoles[i].Username.String < data.UserRoles[j].Username.String
	})

	handlers.TemplateHandler(w, r, "adminWritingsPage.gohtml", data)
}
