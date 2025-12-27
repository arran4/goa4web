package common

import (
	"net/http"

	"github.com/arran4/goa4web/core/consts"
)

// Context keys used by Allowed when reading from the request context.
// These are defined in contextkeys.go.

// Allowed checks if the request context provides one of the given roles.
func Allowed(r *http.Request, roles ...string) bool {
	cd, _ := r.Context().Value(consts.KeyCoreData).(*CoreData)
	if cd == nil {
		return false
	}
	for _, lvl := range roles {
		if lvl == "administrator" {
			if cd.HasAdminAccess() {
				return true
			}
			continue
		}
		if cd.HasRole(lvl) {
			return true
		}
	}
	return false
}
