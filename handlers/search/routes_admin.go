package search

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", remakeCommentsTask.Action).Methods("POST").MatcherFunc(remakeCommentsTask.Match)
	sr.HandleFunc("", remakeNewsTask.Action).Methods("POST").MatcherFunc(remakeNewsTask.Match)
	sr.HandleFunc("", remakeBlogTask.Action).Methods("POST").MatcherFunc(remakeBlogTask.Match)
	sr.HandleFunc("", remakeLinkerTask.Action).Methods("POST").MatcherFunc(remakeLinkerTask.Match)
	sr.HandleFunc("", remakeWritingTask.Action).Methods("POST").MatcherFunc(remakeWritingTask.Match)
	sr.HandleFunc("", remakeImageTask.Action).Methods("POST").MatcherFunc(remakeImageTask.Match)
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
