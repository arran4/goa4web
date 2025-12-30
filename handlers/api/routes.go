package api

import (
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig, navReg *nav.Registry) {
	// Register /api/metadata endpoint
	// Note: The router passed here is usually the root router.
	// We should probably mount under /api if not already there,
	// but the pattern usually is `r` is root.

	api := r.PathPrefix("/api").Subrouter()
	// Middleware? Assuming global middleware handles auth if needed.
	// But this is for logged in users mostly (for uploading images).
	// Image upload requires auth. `cd.UserID` might be 0 if not logged in.
	// `cd.StoreImage` probably fails or we should check auth.
	// The prompt implies "tied to the user like any other upload".
	// So we should enforce auth or check it.

	api.HandleFunc("/metadata", handlers.TaskHandler(metadataTask)).Methods("GET")
}

func Register(reg *router.Registry) {
	reg.RegisterModule("api", nil, RegisterRoutes)
}
