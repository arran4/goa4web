package search

import (
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches the admin search endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sr := ar.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", adminSearchPage).Methods("GET")
	sr.HandleFunc("", tasks.Action(remakeCommentsTask)).Methods("POST").MatcherFunc(remakeCommentsTask.Matcher())
	sr.HandleFunc("", tasks.Action(remakeNewsTask)).Methods("POST").MatcherFunc(remakeNewsTask.Matcher())
	sr.HandleFunc("", tasks.Action(remakeBlogTask)).Methods("POST").MatcherFunc(remakeBlogTask.Matcher())
	sr.HandleFunc("", tasks.Action(remakeLinkerTask)).Methods("POST").MatcherFunc(remakeLinkerTask.Matcher())
	sr.HandleFunc("", tasks.Action(remakeWritingTask)).Methods("POST").MatcherFunc(remakeWritingTask.Matcher())
	sr.HandleFunc("", tasks.Action(remakeImageTask)).Methods("POST").MatcherFunc(remakeImageTask.Matcher())
	sr.HandleFunc("/list", adminSearchWordListPage).Methods("GET")
	sr.HandleFunc("/list.txt", adminSearchWordListDownloadPage).Methods("GET")
}
