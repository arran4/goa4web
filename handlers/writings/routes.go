package writings

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/gorilla/mux"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

var legacyRedirectsEnabled = true

var legacyRedirectsEnabled = true

// RegisterRoutes attaches the public writings endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Writings", "/writings", SectionWeight)
	nav.RegisterAdminControlCenter("Writings", "/admin/writings/categories", SectionWeight)
	wr := r.PathPrefix("/writings").Subrouter()
	wr.Use(handlers.IndexMiddleware(CustomWritingsIndex))
	wr.HandleFunc("/rss", RssPage).Methods("GET")
	wr.HandleFunc("/atom", AtomPage).Methods("GET")
	wr.HandleFunc("", Page).Methods("GET")
	wr.HandleFunc("/", Page).Methods("GET")
	wr.HandleFunc("/writer/{username}", WriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", WriterPage).Methods("GET")
	wr.HandleFunc("/writers", WriterListPage).Methods("GET")
	wr.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("administrator"))
	wr.HandleFunc("/users/permissions", userAllowTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userAllowTask.Matcher())
	wr.HandleFunc("/users/permissions", userDisallowTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("administrator")).MatcherFunc(userDisallowTask.Matcher())
	wr.HandleFunc("/article/{article}", ArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", replyTask.Action).Methods("POST").MatcherFunc(replyTask.Matcher())
	wr.HandleFunc("/article/{article}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(editReplyTask.Action)).ServeHTTP).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	wr.HandleFunc("/article/{article}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(cancelTask.Action)).ServeHTTP).Methods("POST").MatcherFunc(cancelTask.Matcher())
	wr.Handle("/article/{article}/edit", RequireWritingAuthor(http.HandlerFunc(ArticleEditPage))).Methods("GET").MatcherFunc(handlers.RequiredAccess("content writer", "administrator"))
	wr.Handle("/article/{article}/edit", RequireWritingAuthor(http.HandlerFunc(updateWritingTask.Action))).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(updateWritingTask.Matcher())
	wr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", CategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", ArticleAddPage).Methods("GET").MatcherFunc(handlers.RequiredAccess("content writer", "administrator"))
	wr.HandleFunc("/category/{category}/add", submitWritingTask.Action).Methods("POST").MatcherFunc(handlers.RequiredAccess("content writer", "administrator")).MatcherFunc(submitWritingTask.Matcher())

	if legacyRedirectsEnabled {
		// legacy redirects
		r.Path("/writing").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
		r.PathPrefix("/writing/").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
	}
}

// Register registers the writings router module.
func Register() {
	router.RegisterModule("writings", nil, RegisterRoutes)
}
