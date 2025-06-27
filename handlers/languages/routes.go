package languages

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/sections"
)

// RegisterAdminRoutes attaches the admin language endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	sections.RegisterAdminControlCenter("Languages", "/admin/languages", SectionWeight)
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	ar.HandleFunc("/languages", adminLanguagesRenamePage).Methods("POST").MatcherFunc(common.TaskMatcher("Rename Language"))
	ar.HandleFunc("/languages", adminLanguagesDeletePage).Methods("POST").MatcherFunc(common.TaskMatcher("Delete Language"))
	ar.HandleFunc("/languages", adminLanguagesCreatePage).Methods("POST").MatcherFunc(common.TaskMatcher("Create Language"))
}
