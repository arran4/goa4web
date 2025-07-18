package forum

import (
	"net/http"

	common "github.com/arran4/goa4web/handlers/common"
)

// adminForumFlaggedPostsPage displays posts flagged for moderator review.
func AdminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{CoreData: r.Context().Value(common.KeyCoreData).(*CoreData)}
	common.TemplateHandler(w, r, "forumFlaggedPostsPage.gohtml", data)
}
