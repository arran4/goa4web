package bookmarks

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	br := r.PathPrefix("/bookmarks").Subrouter()
	br.Use(handlers.IndexMiddleware(bookmarksCustomIndex))
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/mine", MinePage).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", saveTask.Page).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", handlers.TaskHandler(saveTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveTask.Matcher())
	br.HandleFunc("/edit", handlers.TaskHandler(createTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(createTask.Matcher())
}

// Register registers the bookmarks router module.
func Register() {
	router.RegisterModule("bookmarks", nil, RegisterRoutes)
}
