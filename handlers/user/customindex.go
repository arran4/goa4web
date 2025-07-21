package user

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// CustomIndex injects additional index items for user pages. No items are added currently.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = []common.IndexItem{}
}
