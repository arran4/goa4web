package auth

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
)

// RequiredAccess ensures the requestor has one of the provided access levels.
func RequiredAccess(accessLevels ...string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return corecommon.Allowed(request, accessLevels...)
	}
}

// RequiresAnAccount checks that the requester has a valid user session.
func RequiresAnAccount() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		session, err := core.GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)
		return uid != 0
	}
}
