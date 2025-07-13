package bookmarks

import (
	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"
	router "github.com/arran4/goa4web/internal/router"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router) {
	nav.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	br := r.PathPrefix("/bookmarks").Subrouter()
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/mine", MinePage).Methods("GET").MatcherFunc(auth.RequiresAnAccount())
	br.HandleFunc("/edit", EditPage).Methods("GET").MatcherFunc(auth.RequiresAnAccount())
	br.HandleFunc("/edit", EditSaveActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(SaveTask.Match)
	br.HandleFunc("/edit", EditCreateActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(CreateTask.Match)
	br.HandleFunc("/edit", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount())
}

// Register registers the bookmarks router module.
func Register() {
	router.RegisterModule("bookmarks", nil, RegisterRoutes)
}
