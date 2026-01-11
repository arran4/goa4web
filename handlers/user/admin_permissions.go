package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	u := cd.CurrentProfileUser()
	if u == nil {
		log.Printf("permissions page user not found")
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
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
		log.Printf("get permissions: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})
	data.Rows = rows

	AdminUserPermissionsPage.Handle(w, r, data)
}

const AdminUserPermissionsPage handlers.Page = "admin/userPermissionsPage.gohtml"
