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
	navReg.RegisterIndexLinkWithViewPermission("Linker", "/linker", SectionWeight, "linker", "category")
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Linker"), "Linker", "/admin/linker", SectionWeight)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Linker"), "Categories", "/admin/linker/categories", SectionWeight+1)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Linker"), "Links", "/admin/linker/links", SectionWeight+2)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Linker"), "Queue", "/admin/linker/queue", SectionWeight+3)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Linker"), "Add Link", "/admin/linker/add", SectionWeight+4)
	lr := r.PathPrefix("/linker").Subrouter()
	lr.Use(handlers.IndexMiddleware(CustomLinkerIndex), handlers.SectionMiddleware("linker"))
	lr.HandleFunc("/rss", RssPage).Methods("GET")
	lr.HandleFunc("/u/{username}/rss", RssPage).Methods("GET")
	lr.HandleFunc("/atom", AtomPage).Methods("GET")
	lr.HandleFunc("/u/{username}/atom", AtomPage).Methods("GET")
	lr.HandleFunc("", LinkerPage).Methods("GET")
	lr.HandleFunc("/", LinkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}", UserPage).Methods("GET")
	lr.HandleFunc("/linker/{username}/", UserPage).Methods("GET")
	lr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", LinkerCategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", CommentsPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", handlers.TaskHandler(replyTaskEvent)).Methods("POST").MatcherFunc(replyTaskEvent.Matcher())
	lr.Handle("/comments/{link}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(commentEditAction)))).Methods("POST").MatcherFunc(commentEditAction.Matcher())
	lr.Handle("/comments/{link}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(CommentEditActionCancelPage))).Methods("POST")
	lr.Handle("/show/{link}", handlers.EnforceGrant(ShowPage, "linker", "link", "view", "link")).Methods("GET")
	lr.HandleFunc("/suggest", SuggestPage).Methods("GET")
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
