package search

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers/common"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", adminSearchRemakeCommentsSearchPage).Methods("POST").MatcherFunc(common.TaskMatcher("Remake comments search"))
	sr.HandleFunc("", adminSearchRemakeNewsSearchPage).Methods("POST").MatcherFunc(common.TaskMatcher("Remake news search"))
	sr.HandleFunc("", adminSearchRemakeBlogSearchPage).Methods("POST").MatcherFunc(common.TaskMatcher("Remake blog search"))
	sr.HandleFunc("", adminSearchRemakeLinkerSearchPage).Methods("POST").MatcherFunc(common.TaskMatcher("Remake linker search"))
	sr.HandleFunc("", adminSearchRemakeWritingSearchPage).Methods("POST").MatcherFunc(common.TaskMatcher("Remake writing search"))
	sr.HandleFunc("", adminSearchRemakeImageSearchPage).Methods("POST").MatcherFunc(common.TaskMatcher("Remake image search"))
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
