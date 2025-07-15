package forum

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"net/http"
)

// adminForumFlaggedPostsPage displays posts flagged for moderator review.
func AdminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	hcommon.TemplateHandler("forumFlaggedPostsPage.gohtml").ServeHTTP(w, r)
}
