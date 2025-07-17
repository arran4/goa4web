package forum

import (
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func AdminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{CoreData: r.Context().Value(handlers.KeyCoreData).(*CoreData)}
	handlers.TemplateHandler(w, r, "forumModeratorLogsPage.gohtml", data)
}
