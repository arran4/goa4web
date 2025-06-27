package linker

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/internal/sections"
)

// RegisterRoutes attaches the public linker endpoints to r.
func RegisterRoutes(r *mux.Router) {
	sections.RegisterIndexLink("Linker", "/linker", SectionWeight)
	sections.RegisterAdminControlCenter("Linker", "/admin/linker/categories", SectionWeight)
	lr := r.PathPrefix("/linker").Subrouter()
	lr.HandleFunc("/rss", RssPage).Methods("GET")
	lr.HandleFunc("/atom", AtomPage).Methods("GET")
	lr.HandleFunc("", Page).Methods("GET")
	lr.HandleFunc("/linker/{username}", LinkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}/", LinkerPage).Methods("GET")
	lr.HandleFunc("/categories", CategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", CategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", CommentsPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", CommentsReplyPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	lr.HandleFunc("/show/{link}", ShowPage).Methods("GET")
	lr.HandleFunc("/show/{link}", ShowReplyPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskReply))
	lr.HandleFunc("/suggest", SuggestPage).Methods("GET")
	lr.HandleFunc("/suggest", SuggestActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskSuggest))
}
