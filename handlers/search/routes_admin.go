package search

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", RemakeCommentsTask.Action).Methods("POST").MatcherFunc(RemakeCommentsTask.Match)
	sr.HandleFunc("", RemakeNewsTask.Action).Methods("POST").MatcherFunc(RemakeNewsTask.Match)
	sr.HandleFunc("", RemakeBlogTask.Action).Methods("POST").MatcherFunc(RemakeBlogTask.Match)
	sr.HandleFunc("", RemakeLinkerTask.Action).Methods("POST").MatcherFunc(RemakeLinkerTask.Match)
	sr.HandleFunc("", RemakeWritingTask.Action).Methods("POST").MatcherFunc(RemakeWritingTask.Match)
	sr.HandleFunc("", RemakeImageTask.Action).Methods("POST").MatcherFunc(RemakeImageTask.Match)
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
