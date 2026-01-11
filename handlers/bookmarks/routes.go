package bookmarks

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	br := r.PathPrefix("/bookmarks").Subrouter()
	br.NotFoundHandler = http.HandlerFunc(handlers.RenderNotFoundOrLogin)
	br.Use(handlers.IndexMiddleware(bookmarksCustomIndex))
	br.HandleFunc("", BookmarksPage).Methods("GET")
	br.HandleFunc("/mine", MinePage).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", EditPage).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", handlers.TaskHandler(saveTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveTask.Matcher())
	br.HandleFunc("/edit", handlers.TaskHandler(createTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(createTask.Matcher())

}

// Register registers the bookmarks router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("bookmarks", nil, RegisterRoutes)
}
