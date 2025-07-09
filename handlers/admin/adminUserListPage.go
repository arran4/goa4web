package admin

import (
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
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
	if err := templates.RenderTemplate(w, "admin/userList.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
