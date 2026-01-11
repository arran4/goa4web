package forum

import (
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func AdminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Moderator Logs"
	AdminForumModeratorLogsPageTmpl.Handle(w, r, struct{}{})
}

const AdminForumModeratorLogsPageTmpl handlers.Page = "admin/forumModeratorLogsPage.gohtml"
