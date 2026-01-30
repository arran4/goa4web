package languages

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterAdminRoutes attaches the admin language endpoints to the router.
func RegisterAdminRoutes(r *mux.Router, navReg *navigation.Registry) {
	navReg.RegisterAdminControlCenter("Languages", "Languages", "/admin/languages", SectionWeight)
	r.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	r.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	r.HandleFunc("/languages/new", adminLanguageNewPage).Methods("GET")
	r.HandleFunc("/languages/new", handlers.TaskHandler(createLanguageTask)).Methods("POST").MatcherFunc(createLanguageTask.Matcher())
	r.HandleFunc("/languages/language/{language}", adminLanguagePage).Methods("GET")
	r.HandleFunc("/languages/language/{language}/edit", adminLanguageEditPage).Methods("GET")
	r.HandleFunc("/languages/language/{language}/edit", handlers.TaskHandler(renameLanguageTask)).Methods("POST").MatcherFunc(renameLanguageTask.Matcher())
	r.HandleFunc("/languages/language/{language}/edit", handlers.TaskHandler(deleteLanguageTask)).Methods("POST").MatcherFunc(deleteLanguageTask.Matcher())
}

// Register registers the languages router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("languages", nil, nil)
}
