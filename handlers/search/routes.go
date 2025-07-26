package search

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the search endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) {
	nav.RegisterIndexLink("Search", "/search", SectionWeight)
	nav.RegisterAdminControlCenter("Search", "/admin/search", SectionWeight)
	sr := r.PathPrefix("/search").Subrouter()
	sr.Use(handlers.IndexMiddleware(CustomIndex))
	sr.HandleFunc("", Page).Methods("GET")
	sr.HandleFunc("", handlers.TaskHandler(searchForumTask)).Methods("POST").MatcherFunc(searchForumTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchNewsTask)).Methods("POST").MatcherFunc(searchNewsTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchLinkerTask)).Methods("POST").MatcherFunc(searchLinkerTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchBlogsTask)).Methods("POST").MatcherFunc(searchBlogsTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchWritingsTask)).Methods("POST").MatcherFunc(searchWritingsTask.Matcher())
}

// Register registers the search router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("search", []string{"news"}, RegisterRoutes)
}
