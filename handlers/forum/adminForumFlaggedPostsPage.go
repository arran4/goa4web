package forum

import (
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

// adminForumFlaggedPostsPage displays posts flagged for moderator review.
func AdminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Flagged Posts"
	AdminForumFlaggedPostsPageTmpl.Handle(w, r, struct{}{})
}

const AdminForumFlaggedPostsPageTmpl handlers.Page = "admin/forumFlaggedPostsPage.gohtml"
