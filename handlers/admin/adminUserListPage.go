package admin

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func adminUserListPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	users, err := queries.AllUsers(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Users []*db.AllUsersRow
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Users:    users,
	}
	handlers.TemplateHandler(w, r, "admin/userList.gohtml", data)
}
