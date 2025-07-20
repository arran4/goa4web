package forum

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
)

// adminForumFlaggedPostsPage displays posts flagged for moderator review.
func AdminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}
	data := Data{CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData)}
	handlers.TemplateHandler(w, r, "forumFlaggedPostsPage.gohtml", data)
}
