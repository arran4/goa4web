package languages

import (
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"

	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterAdminRoutes attaches the admin language endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	nav.RegisterAdminControlCenter("Languages", "/admin/languages", SectionWeight)
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	ar.HandleFunc("/languages", tasks.Action(renameLanguageTask)).Methods("POST").MatcherFunc(renameLanguageTask.Matcher())
	ar.HandleFunc("/languages", tasks.Action(deleteLanguageTask)).Methods("POST").MatcherFunc(deleteLanguageTask.Matcher())
	ar.HandleFunc("/languages", tasks.Action(createLanguageTask)).Methods("POST").MatcherFunc(createLanguageTask.Matcher())
}

// Register registers the languages router module.
func Register() {
	router.RegisterModule("languages", nil, func(r *mux.Router) {
		ar := r.PathPrefix("/admin").Subrouter()
		ar.Use(router.AdminCheckerMiddleware)
		RegisterAdminRoutes(ar)
	})
}
