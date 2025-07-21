package router

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
)

// RegisterRoutes sets up all application routes on r.
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/main.css", handlers.MainCSS).Methods("GET")
	r.HandleFunc("/favicon.svg", handlers.Favicon).Methods("GET")

	InitModules(r)

}

// RoleCheckerMiddleware ensures the user has one of the supplied roles.
func RoleCheckerMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !common.Allowed(r, roles...) {
				cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
				err := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", cd)
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
// Roles are loaded via the GetPermissionsByUserID query before this check.
func AdminCheckerMiddleware(next http.Handler) http.Handler {
	return RoleCheckerMiddleware("administrator")(next)
}
