package languages

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterAdminRoutes attaches the admin language endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router, navReg *navpkg.Registry) {
	navReg.RegisterAdminControlCenter("Languages", "/admin/languages", SectionWeight)
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	ar.HandleFunc("/languages", handlers.TaskHandler(renameLanguageTask)).Methods("POST").MatcherFunc(renameLanguageTask.Matcher())
	ar.HandleFunc("/languages", handlers.TaskHandler(deleteLanguageTask)).Methods("POST").MatcherFunc(deleteLanguageTask.Matcher())
	ar.HandleFunc("/languages", handlers.TaskHandler(createLanguageTask)).Methods("POST").MatcherFunc(createLanguageTask.Matcher())
}

// Register registers the languages router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("languages", nil, func(r *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry) {
		ar := r.PathPrefix("/admin").Subrouter()
		ar.Use(router.AdminCheckerMiddleware)
		RegisterAdminRoutes(ar, navReg)
	})
}
