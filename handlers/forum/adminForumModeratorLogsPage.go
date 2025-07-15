package forum

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"net/http"
)

// adminForumModeratorLogsPage displays recent moderator actions.
func AdminForumModeratorLogsPage(w http.ResponseWriter, r *http.Request) {
	hcommon.TemplateHandler("forumModeratorLogsPage.gohtml").ServeHTTP(w, r)
}
