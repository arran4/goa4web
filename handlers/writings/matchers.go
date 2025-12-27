package writings

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// MatchCanEditWritingArticle matches requests where the user may edit the referenced writing article.
func MatchCanEditWritingArticle(r *http.Request, rm *mux.RouteMatch) bool {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		return false
	}
	cd.LoadSelectionsFromRequest(r)
	writingID, err := strconv.Atoi(mux.Vars(r)["writing"])
	if err != nil {
		return false
	}
	return cd.HasGrant("writing", "article", "edit", int32(writingID))
}

// MatchCanPostWriting matches requests where the user may create an article in the referenced category.
func MatchCanPostWriting(r *http.Request, rm *mux.RouteMatch) bool {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		return false
	}
	cd.LoadSelectionsFromRequest(r)
	categoryID, err := strconv.Atoi(mux.Vars(r)["category"])
	if err != nil {
		return false
	}
	return cd.HasGrant("writing", "category", "post", int32(categoryID))
}

// RequireWritingAuthor ensures the requester may edit the writing referenced in the URL.
func RequireWritingAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok {
			http.NotFound(w, r)
			return
		}
		cd.LoadSelectionsFromRequest(r)
		writing, err := cd.CurrentWriting()
		if err != nil {
			log.Printf("RequireWritingAuthor load writing: %v", err)
			http.NotFound(w, r)
			return
		}
		if writing == nil {
			http.NotFound(w, r)
			return
		}
		if !cd.HasGrant("writing", "article", "edit", writing.Idwriting) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
