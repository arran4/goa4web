package router

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/permissions"
	handlers "github.com/arran4/goa4web/pkg/handlers"
)

// RegisterRoutes sets up all application routes on r.
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/main.css", handlers.MainCSS).Methods("GET")
	r.HandleFunc("/favicon.svg", handlers.Favicon).Methods("GET")

	InitModules(r)

	// legacy redirects
	r.Path("/writing").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
	r.PathPrefix("/writing/").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
	r.Path("/links").HandlerFunc(handlers.RedirectPermanentPrefix("/links", "/linker"))
	r.PathPrefix("/links/").HandlerFunc(handlers.RedirectPermanentPrefix("/links", "/linker"))
}

// RoleCheckerMiddleware ensures the user has one of the supplied roles.
func RoleCheckerMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !permissions.Allowed(r, roles...) {
				err := templates.GetCompiledTemplates(corecommon.NewFuncs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData))
				if err != nil {
					log.Printf("Template Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AdminCheckerMiddleware ensures the requester has administrator rights.
func AdminCheckerMiddleware(next http.Handler) http.Handler {
	return RoleCheckerMiddleware("administrator")(next)
}
