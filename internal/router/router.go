package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes sets up all application routes on r.
func RegisterRoutes(r *mux.Router, reg *Registry, cfg *config.RuntimeConfig, navReg *nav.Registry) {
	r.HandleFunc("/robots.txt", handlers.RobotsTXT(cfg)).Methods("GET")
	r.HandleFunc("/main.css", handlers.MainCSS(cfg)).Methods("GET")
	r.HandleFunc("/favicon.svg", handlers.Favicon(cfg)).Methods("GET")
	r.HandleFunc("/static/site.js", handlers.SiteJS(cfg)).Methods("GET")
	r.HandleFunc("/static/a4code.js", handlers.A4CodeJS(cfg)).Methods("GET")

	reg.InitModules(r, cfg, navReg)
}

// RoleCheckerMiddleware ensures the user has one of the supplied roles.
func RoleCheckerMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !common.Allowed(r, roles...) {
				handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
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
