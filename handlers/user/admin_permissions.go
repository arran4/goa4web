package user

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserPermissionsPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	type Data struct {
		*common.CoreData
		User  *db.User
		Rows  []*db.GetPermissionsByUserIDRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := data.Queries()

	if u, err := queries.GetUserById(r.Context(), int32(id)); err == nil {
		data.User = &db.User{Idusers: u.Idusers, Username: u.Username}
	} else {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows, err := queries.GetPermissionsByUserID(r.Context(), int32(id))
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
