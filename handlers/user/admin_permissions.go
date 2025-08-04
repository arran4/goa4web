package user

import (
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	user := cd.CurrentProfileUser()
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	type Data struct {
		*common.CoreData
		User  *db.User
		Rows  []*db.GetPermissionsByUserIDRow
		Roles []*db.Role
	}

	data := Data{CoreData: cd, User: &db.User{Idusers: user.Idusers, Username: user.Username}}

	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows := cd.CurrentProfileRoles()
	if rows == nil {
		rows = []*db.GetPermissionsByUserIDRow{}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})
	data.Rows = rows

	handlers.TemplateHandler(w, r, "userPermissionsPage.gohtml", data)
}
