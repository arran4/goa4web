package router

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes sets up all application routes on r.
func RegisterRoutes(r *mux.Router, reg *Registry, cfg *config.RuntimeConfig, navReg *nav.Registry) {
	r.HandleFunc("/main.css", handlers.MainCSS).Methods("GET")
	r.HandleFunc("/favicon.svg", handlers.Favicon).Methods("GET")

	reg.InitModules(r, cfg, navReg)

}

// RoleCheckerMiddleware ensures the user has one of the supplied roles.
func RoleCheckerMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !common.Allowed(r, roles...) {
				cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
				err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{})
				if err != nil {
					log.Printf("Template Error: %s", err)
					handlers.RenderErrorPage(w, r, err)
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
