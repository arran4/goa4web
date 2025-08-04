package user

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	u := cd.CurrentProfileUser()
	if u == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	id := u.Idusers

	type Data struct {
		User  *db.User
		Rows  []*db.GetPermissionsByUserIDRow
		Roles []*db.Role
	}

	data := Data{User: &db.User{Idusers: u.Idusers, Username: u.Username}}
	queries := cd.Queries()

	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows, err := queries.GetPermissionsByUserID(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})
	data.Rows = rows

	handlers.TemplateHandler(w, r, "userPermissionsPage.gohtml", data)
}
