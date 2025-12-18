package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/gorilla/mux"
)

// TopicPage displays a private topic with thread listings.
func TopicPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topicId := vars["topic"]
	items := []common.IndexItem{
		{
			Name: "Go back to Private Forum",
			Link: "/private",
		},
		{
			Name: "Create a new private thread",
			Link: "/private/topic/" + topicId + "/thread",
		},
	}
	cd.PageCustomIndexItems = append(cd.PageCustomIndexItems, items...)
	forumhandlers.TopicsPageWithBasePath(w, r, "/private")
}
