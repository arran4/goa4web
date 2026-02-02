package languages

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterAdminRoutes attaches the admin language endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router, navReg *navpkg.Registry) {
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Languages"), "Languages", "/admin/languages", SectionWeight)
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	ar.HandleFunc("/languages/new", adminLanguageNewPage).Methods("GET")
	ar.HandleFunc("/languages/new", handlers.TaskHandler(createLanguageTask)).Methods("POST").MatcherFunc(createLanguageTask.Matcher())
	ar.HandleFunc("/languages/language/{language}", adminLanguagePage).Methods("GET")
	ar.HandleFunc("/languages/language/{language}/edit", adminLanguageEditPage).Methods("GET")
	ar.HandleFunc("/languages/language/{language}/edit", handlers.TaskHandler(renameLanguageTask)).Methods("POST").MatcherFunc(renameLanguageTask.Matcher())
	ar.HandleFunc("/languages/language/{language}/edit", handlers.TaskHandler(deleteLanguageTask)).Methods("POST").MatcherFunc(deleteLanguageTask.Matcher())
}

// Register registers the languages router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("languages", nil, nil)
}
