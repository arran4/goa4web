package forum

import (
	"net/http"

	common "github.com/arran4/goa4web/handlers/common"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func AdminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData)}
	common.TemplateHandler(w, r, "forumModeratorLogsPage.gohtml", data)
}
