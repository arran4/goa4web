package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func adminUserListPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	users, err := queries.AllUsers(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Users []*db.AllUsersRow
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Users:    users,
	}
	handlers.TemplateHandler(w, r, "userList.gohtml", data)
}
