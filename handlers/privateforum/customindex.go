package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/gorilla/mux"
)

// CustomIndex injects private forum specific index items.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	vars := mux.Vars(r)
	topicID := vars["topic"]
	items := []common.IndexItem{}
	if topicID == "" {
		items = []common.IndexItem{{
			Name: "Create New private topic",
			Link: "/private/topic/new",
		}}
	} else {
		items = append(items, common.IndexItem{
			Name: "Go back to Private Forum",
			Link: "/private",
		})
	}
	items = append(items, forumhandlers.ForumCustomIndexItems(cd, r, "privateforum")...)
	cd.CustomIndexItems = items
}
