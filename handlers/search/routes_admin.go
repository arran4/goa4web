package search

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", adminSearchRemakeCommentsSearchPage).Methods("POST").MatcherFunc(RemakeCommentsTask.Match)
	sr.HandleFunc("", adminSearchRemakeNewsSearchPage).Methods("POST").MatcherFunc(RemakeNewsTask.Match)
	sr.HandleFunc("", adminSearchRemakeBlogSearchPage).Methods("POST").MatcherFunc(RemakeBlogTask.Match)
	sr.HandleFunc("", adminSearchRemakeLinkerSearchPage).Methods("POST").MatcherFunc(RemakeLinkerTask.Match)
	sr.HandleFunc("", adminSearchRemakeWritingSearchPage).Methods("POST").MatcherFunc(RemakeWritingTask.Match)
	sr.HandleFunc("", adminSearchRemakeImageSearchPage).Methods("POST").MatcherFunc(RemakeImageTask.Match)
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
