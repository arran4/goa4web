package search

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the search endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) []navpkg.RouterOptions {
	opts := []navpkg.RouterOptions{
		navpkg.NewIndexLink("Search", "/search", SectionWeight),
		navpkg.NewAdminControlCenterLink(navpkg.AdminCCCategory("Search"), "Search", "/admin/search", SectionWeight),
	}
	sr := r.PathPrefix("/search").Subrouter()
	sr.Use(handlers.IndexMiddleware(CustomIndex))
	sr.HandleFunc("", SearchPage).Methods("GET")
	sr.HandleFunc("", handlers.TaskHandler(searchForumTask)).Methods("POST").MatcherFunc(searchForumTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchNewsTask)).Methods("POST").MatcherFunc(searchNewsTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchLinkerTask)).Methods("POST").MatcherFunc(searchLinkerTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchBlogsTask)).Methods("POST").MatcherFunc(searchBlogsTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(searchWritingsTask)).Methods("POST").MatcherFunc(searchWritingsTask.Matcher())
	return opts
}

// Register registers the search router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("search", []string{"news"}, func(r *mux.Router, cfg *config.RuntimeConfig) []navpkg.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
