package externallink

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the external link redirect endpoint to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig, _ *nav.Registry, _ *sql.DB, _ sessions.Store) {
	r.HandleFunc("/goto", RedirectHandler).Methods("GET")
	r.HandleFunc("/reload", handlers.TaskHandler(reloadExternalLinkTask)).Methods("POST")
}

// Register registers the external link router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("externallink", nil, RegisterRoutes)
}
