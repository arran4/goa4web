package search

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", remakeCommentsTask.Action).Methods("POST").MatcherFunc(remakeCommentsTask.Matcher())
	sr.HandleFunc("", remakeNewsTask.Action).Methods("POST").MatcherFunc(remakeNewsTask.Matcher())
	sr.HandleFunc("", remakeBlogTask.Action).Methods("POST").MatcherFunc(remakeBlogTask.Matcher())
	sr.HandleFunc("", remakeLinkerTask.Action).Methods("POST").MatcherFunc(remakeLinkerTask.Matcher())
	sr.HandleFunc("", remakeWritingTask.Action).Methods("POST").MatcherFunc(remakeWritingTask.Matcher())
	sr.HandleFunc("", remakeImageTask.Action).Methods("POST").MatcherFunc(remakeImageTask.Matcher())
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
