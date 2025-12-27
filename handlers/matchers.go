package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// RequiredAccess ensures the requestor has one of the provided roles.
func RequiredAccess(accessLevels ...string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return common.Allowed(request, accessLevels...)
	}
}

// RequiredAdminAccess ensures the requestor has administrator grants.
func RequiredAdminAccess() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		cd, _ := request.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil {
			return false
		}
		return cd.HasAdminAccess()
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
