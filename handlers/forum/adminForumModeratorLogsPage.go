package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func AdminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Moderator Logs"
	AdminForumModeratorLogsPageTmpl.Handle(w, r, struct{}{})
}

const AdminForumModeratorLogsPageTmpl tasks.Template = "admin/forumModeratorLogsPage.gohtml"
