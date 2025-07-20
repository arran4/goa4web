package bookmarks

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/gorilla/mux"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	br := r.PathPrefix("/bookmarks").Subrouter()
	br.Use(handlers.IndexMiddleware(func(cd *common.CoreData, r *http.Request) {
		bookmarksCustomIndex(cd)
	}))
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/mine", MinePage).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", saveTask.Page).Methods("GET").MatcherFunc(handlers.RequiresAnAccount())
	br.HandleFunc("/edit", tasks.Action(saveTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(saveTask.Matcher())
	br.HandleFunc("/edit", tasks.Action(createTask)).Methods("POST").MatcherFunc(handlers.RequiresAnAccount()).MatcherFunc(createTask.Matcher())
	br.HandleFunc("/edit", handlers.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(handlers.RequiresAnAccount())
}

// Register registers the bookmarks router module.
func Register() {
	router.RegisterModule("bookmarks", nil, RegisterRoutes)
}
