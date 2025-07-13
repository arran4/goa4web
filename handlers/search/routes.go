package search

import (
	"github.com/gorilla/mux"

	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the search endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Search", "/search", SectionWeight)
	nav.RegisterAdminControlCenter("Search", "/admin/search", SectionWeight)
	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", Page).Methods("GET")
	sr.HandleFunc("", SearchForumTask.Action()).Methods("POST").MatcherFunc(SearchForumTask.Match)
	sr.HandleFunc("", SearchNewsTask.Action()).Methods("POST").MatcherFunc(SearchNewsTask.Match)
	sr.HandleFunc("", SearchLinkerTask.Action()).Methods("POST").MatcherFunc(SearchLinkerTask.Match)
	sr.HandleFunc("", SearchBlogsTask.Action()).Methods("POST").MatcherFunc(SearchBlogsTask.Match)
	sr.HandleFunc("", SearchWritingsTask.Action()).Methods("POST").MatcherFunc(SearchWritingsTask.Match)
}

// Register registers the search router module.
func Register() {
	router.RegisterModule("search", []string{"news"}, RegisterRoutes)
}
