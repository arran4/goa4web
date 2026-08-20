package privateforum

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// DisablePrivateForumCaching prevents browsers and shared caches from storing private forum responses.
func DisablePrivateForumCaching(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.DisableCaching(w)

		next.ServeHTTP(w, r)
	})
}

// RequirePrivateTopicAccess middleware checks for exact-topic view membership.
func RequirePrivateTopicAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		vars := mux.Vars(r)
		topicStr, ok := vars["topic"]
		if !ok || topicStr == "" {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		topicID, err := strconv.Atoi(topicStr)
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		top, err := cd.ForumTopicByID(int32(topicID))
		if err != nil || top == nil {
			handlers.RenderNotFoundOrLogin(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// EnforcePrivateForumTopicSeeAccess middleware checks for see grant on private forum topic 0.
func EnforcePrivateForumTopicSeeAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		if !cd.HasGrant("privateforum", "topic", "see", 0) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
