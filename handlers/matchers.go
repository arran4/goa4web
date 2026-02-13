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
		if cd, ok := request.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
			if cd.IsUserLoggedIn() {
				return true
			}
		}
		session, err := core.GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)
		return uid != 0
	}
}

// RequireGrant ensures the requester holds the specified grant before matching the route.
func RequireGrant(section, item, action string, resolveItemID func(r *http.Request, match *mux.RouteMatch) (int32, bool)) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		cd, ok := request.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			return false
		}
		var itemID int32
		if resolveItemID != nil {
			var ok bool
			itemID, ok = resolveItemID(request, match)
			if !ok {
				return false
			}
		}
		return cd.HasGrant(section, item, action, itemID)
	}
}

// RequireGrantForPathInt checks for a grant tied to an integer path parameter.
func RequireGrantForPathInt(section, item, action, param string) mux.MatcherFunc {
	return RequireGrant(section, item, action, func(r *http.Request, match *mux.RouteMatch) (int32, bool) {
		if match != nil && match.Vars != nil {
			if id, err := strconv.Atoi(match.Vars[param]); err == nil {
				return int32(id), true
			}
		}
		if vars := mux.Vars(r); vars != nil {
			if id, err := strconv.Atoi(vars[param]); err == nil {
				return int32(id), true
			}
		}
		return 0, false
	})
}
