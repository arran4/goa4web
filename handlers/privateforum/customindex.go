package privateforum

import (
	"fmt"
	"net/http"
	"strconv"

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
		if tid, err := strconv.Atoi(topicID); err == nil { // TODO check for post permission
			_ = tid
			items = append(items, common.IndexItem{
				Name: "Create a new private thread",
				Link: fmt.Sprintf("/private/topic/%s/thread", topicID),
			})
		}
	}
	forumItems := forumhandlers.ForumCustomIndexItems(cd, r)
	for _, item := range forumItems {
		if item.Name != "New Thread" {
			items = append(items, item)
		}
	}
	cd.CustomIndexItems = items
}
