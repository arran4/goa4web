package search

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", handlers.TaskHandler(remakeCommentsTask)).Methods("POST").MatcherFunc(remakeCommentsTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(remakeNewsTask)).Methods("POST").MatcherFunc(remakeNewsTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(remakeBlogTask)).Methods("POST").MatcherFunc(remakeBlogTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(remakeLinkerTask)).Methods("POST").MatcherFunc(remakeLinkerTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(remakeWritingTask)).Methods("POST").MatcherFunc(remakeWritingTask.Matcher())
	sr.HandleFunc("", handlers.TaskHandler(remakeImageTask)).Methods("POST").MatcherFunc(remakeImageTask.Matcher())
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
