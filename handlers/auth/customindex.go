package auth

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
)

// CustomIndex injects additional index items for auth pages. Currently it does nothing.
var CustomIndex = func(cd *common.CoreData, r *http.Request) {}
