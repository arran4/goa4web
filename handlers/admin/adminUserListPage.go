package admin

import (
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
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
		*corecommon.CoreData
		Users []*db.AllUsersRow
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Users:    users,
	}
	common.TemplateHandler(w, r, "admin/userList.gohtml", data)
}
