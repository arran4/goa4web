package forum

import (
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
)

// adminForumFlaggedPostsPage displays posts flagged for moderator review.
func AdminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{CoreData: r.Context().Value(handlers.KeyCoreData).(*CoreData)}
	handlers.TemplateHandler(w, r, "forumFlaggedPostsPage.gohtml", data)
}
