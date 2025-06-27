package bookmarks

import (
	"github.com/gorilla/mux"

	auth "github.com/arran4/goa4web/handlers/auth"
	hcommon "github.com/arran4/goa4web/handlers/common"

	"github.com/arran4/goa4web/internal/sections"
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router) {
	sections.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)
	br := r.PathPrefix("/bookmarks").Subrouter()
	br.HandleFunc("", Page).Methods("GET")
	br.HandleFunc("/mine", MinePage).Methods("GET").MatcherFunc(auth.RequiresAnAccount())
	br.HandleFunc("/edit", EditPage).Methods("GET").MatcherFunc(auth.RequiresAnAccount())
	br.HandleFunc("/edit", EditSaveActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSave))
	br.HandleFunc("/edit", EditCreateActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCreate))
	br.HandleFunc("/edit", hcommon.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount())
}
