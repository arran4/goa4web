package linker

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"net/http"

	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

var legacyRedirectsEnabled = true

// RegisterRoutes attaches the public linker endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Linker", "/linker", SectionWeight)
	navReg.RegisterAdminControlCenter("Linker", "/admin/linker/categories", SectionWeight)
	lr := r.PathPrefix("/linker").Subrouter()
	lr.Use(handlers.IndexMiddleware(CustomLinkerIndex))
	lr.HandleFunc("/rss", RssPage).Methods("GET")
	lr.HandleFunc("/atom", AtomPage).Methods("GET")
	lr.HandleFunc("", Page).Methods("GET")
	lr.HandleFunc("/", Page).Methods("GET")
	lr.HandleFunc("/linker/{username}", UserPage).Methods("GET")
	lr.HandleFunc("/linker/{username}/", UserPage).Methods("GET")
	lr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", CategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", replyTaskEvent.Page).Methods("GET")
	lr.HandleFunc("/comments/{link}", handlers.TaskHandler(replyTaskEvent)).Methods("POST").MatcherFunc(replyTaskEvent.Matcher())
	lr.Handle("/comments/{link}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(commentEditAction.Page))).Methods("POST").MatcherFunc(commentEditAction.Matcher())
	lr.Handle("/comments/{link}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(commentEditActionCancel.Page))).Methods("POST").MatcherFunc(commentEditActionCancel.Matcher())
	lr.HandleFunc("/show/{link}", replyTaskEvent.Page).Methods("GET")
	lr.HandleFunc("/suggest", suggestTask.Page).Methods("GET")
	lr.HandleFunc("/suggest", handlers.TaskHandler(suggestTask)).Methods("POST").MatcherFunc(suggestTask.Matcher())

	if legacyRedirectsEnabled {
		// legacy redirects
		r.Path("/links").HandlerFunc(handlers.RedirectPermanentPrefix("/links", "/linker"))
		r.PathPrefix("/links/").HandlerFunc(handlers.RedirectPermanentPrefix("/links", "/linker"))
	}
}

// Register registers the linker router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("linker", nil, RegisterRoutes)
}
