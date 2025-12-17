package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// CustomIndex injects private forum specific index items.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = []common.IndexItem{}
	// Action to start a new private group discussion
	cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
		Name: "new group topic",
		Link: "/private/topic/new",
	})
}
