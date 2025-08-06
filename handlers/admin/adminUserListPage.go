package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserListPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Users"
	queries := cd.Queries()
	users, err := queries.AdminListAllUsers(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := struct {
		Users []*db.AdminListAllUsersRow
	}{
		Users: users,
	}
	handlers.TemplateHandler(w, r, "userList.gohtml", data)
}
