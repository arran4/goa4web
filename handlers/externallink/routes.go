package externallink

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the external link redirect endpoint to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) []nav.RouterOptions {
	r.HandleFunc("/goto", RedirectHandler).Methods("GET")
	r.HandleFunc("/goto", handlers.TaskHandler(reloadExternalLinkTask)).Methods("POST")
	return nil
}

// Register registers the external link router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("externallink", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []nav.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
