package admin

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
)

// CustomIndex injects additional index items for admin pages. Currently it is empty.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = []common.IndexItem{}
}
