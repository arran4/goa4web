package images

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
)

// CustomIndex injects additional index items for image routes. No custom items are provided.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {}
