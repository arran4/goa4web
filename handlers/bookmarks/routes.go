package bookmarks

import (
	"database/sql"
	"net/http"

	"github.com/arran4/gobookmarks"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	// "github.com/arran4/goa4web/handlers" // Removed as we replace handlers
	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterRoutes attaches the bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry, db *sql.DB, store sessions.Store) {
	navReg.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)

	provider := &GoBookmarksUserProvider{}

	// Initialize gobookmarks configuration
	gbCfg := &gobookmarks.RouterConfig{
		DB:           db,
		UserProvider: provider,
		SessionStore: store,
		SessionName:  cfg.SessionName,
		// defaults...
		BaseURL:     "/bookmarks",
		ExternalURL: cfg.HTTPHostname + "/bookmarks",
		DevMode:     false, // Or verify via config
	}

	router := gobookmarks.NewRouter(gbCfg)

	// Mount it
	r.PathPrefix("/bookmarks").Handler(http.StripPrefix("/bookmarks", router))
}

// Register registers the bookmarks router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("bookmarks", nil, RegisterRoutes)
}
