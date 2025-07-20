package forum

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func AdminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}
	data := Data{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	handlers.TemplateHandler(w, r, "forumModeratorLogsPage.gohtml", data)
}
