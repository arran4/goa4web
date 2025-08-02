package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserListPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Users"
	queries := cd.Queries()
	users, err := queries.AdminAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Users []*db.AdminAllUsersRow
	}{
		CoreData: cd,
		Users:    users,
	}
	handlers.TemplateHandler(w, r, "userList.gohtml", data)
}
