package search

import (
	"github.com/gorilla/mux"

	news "github.com/arran4/goa4web/handlers/news"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the search endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Search", "/search", SectionWeight)
	nav.RegisterAdminControlCenter("Search", "/admin/search", SectionWeight)
	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", Page).Methods("GET")
	sr.HandleFunc("", SearchResultForumActionPage).Methods("POST").MatcherFunc(SearchForumTask.Match)
	sr.HandleFunc("", news.SearchResultNewsActionPage).Methods("POST").MatcherFunc(SearchNewsTask.Match)
	sr.HandleFunc("", SearchResultLinkerActionPage).Methods("POST").MatcherFunc(SearchLinkerTask.Match)
	sr.HandleFunc("", SearchResultBlogsActionPage).Methods("POST").MatcherFunc(SearchBlogsTask.Match)
	sr.HandleFunc("", SearchResultWritingsActionPage).Methods("POST").MatcherFunc(SearchWritingsTask.Match)
}

// Register registers the search router module.
func Register() {
	router.RegisterModule("search", []string{"news"}, RegisterRoutes)
}
