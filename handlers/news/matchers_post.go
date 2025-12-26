package news

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// MatchCanPostNews matches requests where the user has permission to post news.
func MatchCanPostNews(r *http.Request, rm *mux.RouteMatch) bool {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		return false
	}
	return CanPostNews(cd)
}
