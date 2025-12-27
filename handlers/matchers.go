package handlers

import (
	"net/http"
	"strconv"

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

// RequiredGrant ensures the requestor has the specified grant.
func RequiredGrant(section, item, action string, itemID int32) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		cd, _ := request.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil {
			return false
		}
		return cd.HasGrant(section, item, action, itemID)
	}
}

// RequiredGrantFromPath ensures the requestor has the grant for the specified item ID stored in the URL path.
func RequiredGrantFromPath(section, item, action, param string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		cd, _ := request.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil {
			return false
		}
		vars := mux.Vars(request)
		val := vars[param]
		if val == "" {
			return false
		}
		id, err := strconv.Atoi(val)
		if err != nil {
			return false
		}
		return cd.HasGrant(section, item, action, int32(id))
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
