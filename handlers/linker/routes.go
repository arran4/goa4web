package linker

import (
	"net/http"

	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/arran4/goa4web/internal/tasks"

	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/gorilla/mux"

	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// AddLinkerIndex injects linker index links into CoreData.
func AddLinkerIndex(h http.Handler) http.Handler {
	return hcommon.IndexMiddleware(CustomLinkerIndex)(h)
}

var legacyRedirectsEnabled = true

// RegisterRoutes attaches the public linker endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Linker", "/linker", SectionWeight)
	nav.RegisterAdminControlCenter("Linker", "/admin/linker/categories", SectionWeight)
	lr := r.PathPrefix("/linker").Subrouter()
	lr.Use(hcommon.IndexMiddleware(CustomLinkerIndex))
	lr.HandleFunc("/rss", RssPage).Methods("GET")
	lr.HandleFunc("/atom", AtomPage).Methods("GET")
	lr.HandleFunc("", Page).Methods("GET")
	lr.HandleFunc("/linker/{username}", LinkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}/", LinkerPage).Methods("GET")
	lr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", CategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", CommentsPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", ReplyTaskEvent.Action).Methods("POST").MatcherFunc(ReplyTaskEvent.Match)
	lr.Handle("/comments/{link}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(CommentEditActionPage))).Methods("POST").MatcherFunc(EditReplyTask.Match)
	lr.Handle("/comments/{link}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(CommentEditActionCancelPage))).Methods("POST").MatcherFunc(tasks.CancelTask.Match)
	lr.HandleFunc("/show/{link}", ShowPage).Methods("GET")
	lr.HandleFunc("/show/{link}", ReplyTaskEvent.Action).Methods("POST").MatcherFunc(ReplyTaskEvent.Match)
	lr.HandleFunc("/suggest", SuggestPage).Methods("GET")
	lr.HandleFunc("/suggest", SuggestTask.Action).Methods("POST").MatcherFunc(SuggestTask.Match)

	if legacyRedirectsEnabled {
		// legacy redirects
		r.Path("/links").HandlerFunc(hcommon.RedirectPermanentPrefix("/links", "/linker"))
		r.PathPrefix("/links/").HandlerFunc(hcommon.RedirectPermanentPrefix("/links", "/linker"))
	}
}

// Register registers the linker router module.
func Register() {
	router.RegisterModule("linker", nil, RegisterRoutes)
}
