package search

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// CustomIndex injects additional index items for search pages. It is empty for now.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = []common.IndexItem{}
}
