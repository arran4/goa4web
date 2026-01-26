package forum

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
)

// adminForumFlaggedPostsPage displays posts flagged for moderator review.
func AdminForumFlaggedPostsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Flagged Posts"
	AdminForumFlaggedPostsPageTmpl.Handle(w, r, struct{}{})
}

const AdminForumFlaggedPostsPageTmpl tasks.Template = "admin/forumFlaggedPostsPage.gohtml"
