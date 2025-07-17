package bookmarks

import (
	"net/http"

	"github.com/gorilla/mux"

	handlers "github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// AddBookmarksIndex injects bookmark index links into CoreData.
func AddBookmarksIndex(h http.Handler) http.Handler {
	return handlers.IndexMiddleware(func(cd *handlers.CoreData, r *http.Request) {
		bookmarksCustomIndex(cd)
	})(h)
}

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	br := r.PathPrefix("/bookmarks").Subrouter()
	br.Use(AddBookmarksIndex)
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/mine", MinePage).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", saveTask.Page).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", saveTask.Action).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveTask.Match)
	br.HandleFunc("/edit", createTask.Action).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(createTask.Match)
	br.HandleFunc("/edit", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(handlers.RequiresAnAccount())
}

// Register registers the bookmarks router module.
func Register() {
	router.RegisterModule("bookmarks", nil, RegisterRoutes)
}
