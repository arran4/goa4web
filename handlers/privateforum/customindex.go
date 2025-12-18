package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
)

// CustomIndex injects private forum specific index items.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	items := []common.IndexItem{{
		Name: "Start Group Discussion",
		Link: "/private/topic/new",
	}}
	cd.CustomIndexItems = append(items, forumhandlers.ForumCustomIndexItems(cd, r)...)
}
