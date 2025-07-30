package externallink

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the external link redirect endpoint to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, _ *nav.Registry) {
	r.HandleFunc("/goto", RedirectHandler).Methods("GET")
}

// Register registers the external link router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("externallink", nil, RegisterRoutes)
}
