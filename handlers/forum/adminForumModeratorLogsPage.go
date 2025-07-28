package forum

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func AdminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Moderator Logs"
	data := Data{CoreData: cd}
	handlers.TemplateHandler(w, r, "forumModeratorLogsPage.gohtml", data)
}
