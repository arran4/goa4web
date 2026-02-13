package admin

import (
	"database/sql"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/gorilla/mux"
)

// Handlers bundles dependencies used by admin handlers.
type Handlers struct {
	ConfigFile          string
	Srv                 *server.Server
	DBPool              *sql.DB
	UpdateConfigKeyFunc func(fs core.FileSystem, path, key, value string) error
}

// Option configures Handlers returned by New.
type Option func(*Handlers)

// New creates a Handlers value configured by opts.
func New(opts ...Option) *Handlers {
	h := &Handlers{}
	for _, o := range opts {
		o(h)
	}
	return h
}

// WithConfigFile sets the configuration file path.
func WithConfigFile(path string) Option { return func(h *Handlers) { h.ConfigFile = path } }

// WithServer sets the server instance.
func WithServer(s *server.Server) Option { return func(h *Handlers) { h.Srv = s } }

// WithDBPool sets the database pool.
func WithDBPool(db *sql.DB) Option { return func(h *Handlers) { h.DBPool = db } }

// WithUpdateConfigKeyFunc sets the configuration update function.
func WithUpdateConfigKeyFunc(fn func(fs core.FileSystem, path, key, value string) error) Option {
	return func(h *Handlers) { h.UpdateConfigKeyFunc = fn }
}

// Register registers the admin router module using h's dependencies.
func (h *Handlers) Register(reg *router.Registry) {
	reg.RegisterModule("admin", []string{"faq", "forum", "imagebbs", "languages", "linker", "news", "search", "user", "writings", "blogs"}, func(r *mux.Router, cfg *config.RuntimeConfig) []navigation.RouterOptions {
		ar := r.PathPrefix("/admin").Subrouter()
		ar.Use(router.AdminCheckerMiddleware)
		ar.Use(handlers.IndexMiddleware(CustomIndex))
		return h.RegisterRoutes(ar, cfg)
	})
}
