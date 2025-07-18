package languages

import (
	"github.com/gorilla/mux"

	nav "github.com/arran4/goa4web/internal/navigation"
	router "github.com/arran4/goa4web/internal/router"
)

// RegisterAdminRoutes attaches the admin language endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	nav.RegisterAdminControlCenter("Languages", "/admin/languages", SectionWeight)
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	ar.HandleFunc("/languages", RenameLanguageTask.Action).Methods("POST").MatcherFunc(RenameLanguageTask.Match)
	ar.HandleFunc("/languages", DeleteLanguageTask.Action).Methods("POST").MatcherFunc(DeleteLanguageTask.Match)
	ar.HandleFunc("/languages", CreateLanguageTask.Action).Methods("POST").MatcherFunc(CreateLanguageTask.Match)
}

// Register registers the languages router module.
func Register() {
	router.RegisterModule("languages", nil, func(r *mux.Router) {
		ar := r.PathPrefix("/admin").Subrouter()
		ar.Use(router.AdminCheckerMiddleware)
		RegisterAdminRoutes(ar)
	})
}
