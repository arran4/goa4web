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
	sr.HandleFunc("", searchForumTask.Action).Methods("POST").MatcherFunc(searchForumTask.Match)
	sr.HandleFunc("", searchNewsTask.Action).Methods("POST").MatcherFunc(searchNewsTask.Match)
	sr.HandleFunc("", searchLinkerTask.Action).Methods("POST").MatcherFunc(searchLinkerTask.Match)
	sr.HandleFunc("", searchBlogsTask.Action).Methods("POST").MatcherFunc(searchBlogsTask.Match)
	sr.HandleFunc("", searchWritingsTask.Action).Methods("POST").MatcherFunc(searchWritingsTask.Match)
}

// Register registers the search router module.
func Register() {
	router.RegisterModule("search", []string{"news"}, RegisterRoutes)
}
