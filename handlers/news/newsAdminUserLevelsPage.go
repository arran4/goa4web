package news

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminUserRolesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		UserLevels []*db.GetUserRolesRow
		Roles      []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.CoreData.PageTitle = "News Roles"

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
	data.UserLevels = rows

	handlers.TemplateHandler(w, r, "adminUserRolesPage.gohtml", data)
}
