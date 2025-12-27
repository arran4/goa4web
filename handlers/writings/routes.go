package writings

import (
	"github.com/arran4/goa4web/handlers/forum/comments"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

var legacyRedirectsEnabled = true

// RegisterRoutes attaches the public writings endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLinkWithViewPermission("Writings", "/writings", SectionWeight, "writing", "category")
	navReg.RegisterAdminControlCenter("Writings", "Writings", "/admin/writings", SectionWeight)
	wr := r.PathPrefix("/writings").Subrouter()
	wr.Use(handlers.IndexMiddleware(CustomWritingsIndex), handlers.SectionMiddleware("writing"))
	wr.HandleFunc("/rss", RssPage).Methods("GET")
	wr.HandleFunc("/u/{username}/rss", RssPage).Methods("GET")
	wr.HandleFunc("/atom", AtomPage).Methods("GET")
	wr.HandleFunc("/u/{username}/atom", AtomPage).Methods("GET")
	wr.HandleFunc("", WritingsPage).Methods("GET")
	wr.HandleFunc("/", WritingsPage).Methods("GET")
	wr.HandleFunc("/writer/{username}", WriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", WriterPage).Methods("GET")
	wr.HandleFunc("/writers", WriterListPage).Methods("GET")
	// Writing routes use {writing} to identify the requested writing.
	wr.HandleFunc("/article/{writing}", ArticlePage).Methods("GET")
	wr.HandleFunc("/article/{writing}", handlers.TaskHandler(replyTask)).Methods("POST").MatcherFunc(replyTask.Matcher())
	wr.Handle("/article/{writing}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(editReplyTask)))).Methods("POST").MatcherFunc(editReplyTask.Matcher())
	wr.Handle("/article/{writing}/comment/{comment}", comments.RequireCommentAuthor(http.HandlerFunc(handlers.TaskHandler(cancelTask)))).Methods("POST").MatcherFunc(cancelTask.Matcher())
	wr.Handle("/article/{writing}/edit", RequireWritingAuthor(http.HandlerFunc(updateWritingTask.Page))).Methods("GET").MatcherFunc(MatchCanEditWritingArticle)
	wr.Handle("/article/{writing}/edit", RequireWritingAuthor(http.HandlerFunc(handlers.TaskHandler(updateWritingTask)))).Methods("POST").MatcherFunc(MatchCanEditWritingArticle).MatcherFunc(updateWritingTask.Matcher())
	wr.HandleFunc("/article/{writing}/labels", handlers.TaskHandler(setLabelsTask)).Methods("POST").MatcherFunc(setLabelsTask.Matcher())
	wr.HandleFunc("/article/{writing}/labels", handlers.TaskHandler(markWritingReadTask)).Methods("GET", "POST").MatcherFunc(markWritingReadTask.Matcher())
	wr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	wr.HandleFunc("/categories/", CategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", CategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", submitWritingTask.Page).Methods("GET").MatcherFunc(MatchCanPostWriting)
	wr.HandleFunc("/category/{category}/add", handlers.TaskHandler(submitWritingTask)).Methods("POST").MatcherFunc(MatchCanPostWriting).MatcherFunc(submitWritingTask.Matcher())

	if legacyRedirectsEnabled {
		// legacy redirects
		r.Path("/writing").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
		r.PathPrefix("/writing/").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
	}
}

// Register registers the writings router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("writings", nil, RegisterRoutes)
}
