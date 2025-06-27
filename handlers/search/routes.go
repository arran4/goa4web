package search

import (
	"github.com/gorilla/mux"

	hcommon "github.com/arran4/goa4web/handlers/common"
	news "github.com/arran4/goa4web/handlers/news"

	"github.com/arran4/goa4web/internal/sections"
)

// RegisterRoutes attaches the search endpoints to r.
func RegisterRoutes(r *mux.Router) {
	sections.RegisterIndexLink("Search", "/search", SectionWeight)
	sections.RegisterAdminControlCenter("Search", "/admin/search", SectionWeight)
	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", Page).Methods("GET")
	sr.HandleFunc("", SearchResultForumActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSearchForum))
	sr.HandleFunc("", news.SearchResultNewsActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSearchNews))
	sr.HandleFunc("", SearchResultLinkerActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSearchLinker))
	sr.HandleFunc("", SearchResultBlogsActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSearchBlogs))
	sr.HandleFunc("", SearchResultWritingsActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSearchWritings))
}
