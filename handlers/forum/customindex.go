package forum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/forumcommon"
)

// CustomForumIndex builds context-aware index items for the public forum.
func CustomForumIndex(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = forumcommon.ForumCustomIndexItems(cd, r, "forum")
}
