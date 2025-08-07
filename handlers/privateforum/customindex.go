package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// CustomIndex injects private forum specific index items. None are added yet.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {
	cd.CustomIndexItems = []common.IndexItem{}
}
