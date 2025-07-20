package search

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"

	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the search endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Search", "/search", SectionWeight)
	nav.RegisterAdminControlCenter("Search", "/admin/search", SectionWeight)
	sr := r.PathPrefix("/search").Subrouter()
	sr.Use(handlers.IndexMiddleware(CustomIndex))
	sr.HandleFunc("", Page).Methods("GET")
	sr.HandleFunc("", tasks.Action(searchForumTask)).Methods("POST").MatcherFunc(searchForumTask.Matcher())
	sr.HandleFunc("", tasks.Action(searchNewsTask)).Methods("POST").MatcherFunc(searchNewsTask.Matcher())
	sr.HandleFunc("", tasks.Action(searchLinkerTask)).Methods("POST").MatcherFunc(searchLinkerTask.Matcher())
	sr.HandleFunc("", tasks.Action(searchBlogsTask)).Methods("POST").MatcherFunc(searchBlogsTask.Matcher())
	sr.HandleFunc("", tasks.Action(searchWritingsTask)).Methods("POST").MatcherFunc(searchWritingsTask.Matcher())
}

// Register registers the search router module.
func Register() {
	router.RegisterModule("search", []string{"news"}, RegisterRoutes)
}
