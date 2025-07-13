package search

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", adminSearchRemakeCommentsSearchPage).Methods("POST").MatcherFunc(RemakeCommentsTask.Matcher)
	sr.HandleFunc("", adminSearchRemakeNewsSearchPage).Methods("POST").MatcherFunc(RemakeNewsTask.Matcher)
	sr.HandleFunc("", adminSearchRemakeBlogSearchPage).Methods("POST").MatcherFunc(RemakeBlogTask.Matcher)
	sr.HandleFunc("", adminSearchRemakeLinkerSearchPage).Methods("POST").MatcherFunc(RemakeLinkerTask.Matcher)
	sr.HandleFunc("", adminSearchRemakeWritingSearchPage).Methods("POST").MatcherFunc(RemakeWritingTask.Matcher)
	sr.HandleFunc("", adminSearchRemakeImageSearchPage).Methods("POST").MatcherFunc(RemakeImageTask.Matcher)
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
