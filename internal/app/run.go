package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/workers"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	corelanguage "github.com/arran4/goa4web/core/language"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dlq"
	email "github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	imagesign "github.com/arran4/goa4web/internal/images"
	middleware "github.com/arran4/goa4web/internal/middleware"
	csrfmw "github.com/arran4/goa4web/internal/middleware/csrf"
	nav "github.com/arran4/goa4web/internal/navigation"
	routerpkg "github.com/arran4/goa4web/internal/router"
	websocket "github.com/arran4/goa4web/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// ConfigFile stores the path to the configuration file if provided on the
// command line. It is used by admin handlers when updating settings.
var ConfigFile string

var (
	version = "dev"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

// ServerOption configures additional parameters when constructing a server.
type ServerOption func(*serverOptions)

type serverOptions struct {
	SessionSecret   string
	ImageSignSecret string
	APISecret       string
	DBReg           *dbdrivers.Registry
	EmailReg        *email.Registry
	DLQReg          *dlq.Registry
	Bus             *eventbus.Bus
	Store           *sessions.CookieStore
	DB              *sql.DB
	RouterReg       *routerpkg.Registry
}

// WithSessionSecret supplies the session cookie encryption secret.
func WithSessionSecret(secret string) ServerOption {
	return func(o *serverOptions) { o.SessionSecret = secret }
}

// WithImageSignSecret supplies the image signing secret.
func WithImageSignSecret(secret string) ServerOption {
	return func(o *serverOptions) { o.ImageSignSecret = secret }
}

// WithAPISecret sets the administrator API secret.
func WithAPISecret(secret string) ServerOption {
	return func(o *serverOptions) { o.APISecret = secret }
}

// WithDBRegistry sets the database driver registry used to initialise the pool.
func WithDBRegistry(r *dbdrivers.Registry) ServerOption {
	return func(o *serverOptions) { o.DBReg = r }
}

// WithEmailRegistry sets the email provider registry.
func WithEmailRegistry(r *email.Registry) ServerOption {
	return func(o *serverOptions) { o.EmailReg = r }
}

// WithDLQRegistry sets the dead letter queue provider registry.
func WithDLQRegistry(r *dlq.Registry) ServerOption {
	return func(o *serverOptions) { o.DLQReg = r }
}

// WithBus uses the provided event bus instead of creating a new one.
func WithBus(b *eventbus.Bus) ServerOption { return func(o *serverOptions) { o.Bus = b } }

// WithStore uses the supplied session store instead of creating one.
func WithStore(s *sessions.CookieStore) ServerOption { return func(o *serverOptions) { o.Store = s } }

// WithDB uses the supplied database pool instead of performing startup checks.
func WithDB(db *sql.DB) ServerOption { return func(o *serverOptions) { o.DB = db } }

// WithRouterRegistry sets the router module registry.
func WithRouterRegistry(r *routerpkg.Registry) ServerOption {
	return func(o *serverOptions) { o.RouterReg = r }
}

// NewServer constructs the server and supporting services using the provided
// configuration and optional parameters.
func NewServer(ctx context.Context, cfg config.RuntimeConfig, opts ...ServerOption) (*server.Server, error) {
	o := &serverOptions{}
	for _, op := range opts {
		op(o)
	}

	log.Printf("application version %s starting", version)
	adminhandlers.StartTime = time.Now()

	store := o.Store
	if store == nil {
		if o.SessionSecret == "" {
			return nil, fmt.Errorf("session secret required")
		}
		store = sessions.NewCookieStore([]byte(o.SessionSecret))
	}
	core.Store = store
	core.SessionName = cfg.SessionName
	store.Options = &sessions.Options{Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}

	dbPool := o.DB
	if dbPool == nil {
		var err error
		dbPool, err = PerformChecks(cfg, o.DBReg)
		if err != nil {
			return nil, fmt.Errorf("startup checks: %w", err)
		}
	}
	queries := dbpkg.New(dbPool)
	if err := corelanguage.EnsureDefaultLanguage(context.Background(), queries, cfg.DefaultLanguage); err != nil {
		return nil, fmt.Errorf("ensure default language: %w", err)
	}
	if err := corelanguage.ValidateDefaultLanguage(context.Background(), queries, cfg.DefaultLanguage); err != nil {
		return nil, fmt.Errorf("default language: %w", err)
	}

	if err := config.ApplySMTPFallbacks(&cfg); err != nil {
		return nil, fmt.Errorf("smtp fallback: %w", err)
	}
	config.AppRuntimeConfig = cfg
	imgSigner := imagesign.NewSigner(cfg, o.ImageSignSecret)
	adminhandlers.AdminAPISecret = o.APISecret
	email.SetDefaultFromName(cfg.EmailFrom)

	bus := o.Bus
	if bus == nil {
		bus = eventbus.NewBus()
	}
	websocket.SetBus(bus)

	reg := o.RouterReg
	if reg == nil {
		reg = routerpkg.NewRegistry()
	}
	r := mux.NewRouter()
	routerpkg.RegisterRoutes(r, reg)

	srv := server.New(nil, store, dbPool, cfg, reg, navReg, o.DLQReg)
	nav.SetDefaultRegistry(navReg) // TODO make it work like the others.
	srv.Bus = bus
	srv.EmailReg = o.EmailReg
	srv.ImageSigner = imgSigner

	taskEventMW := middleware.NewTaskEventMiddleware(bus)
	handler := middleware.NewMiddlewareChain(
		middleware.RecoverMiddleware,
		srv.CoreDataMiddleware(),
		middleware.RequestLoggerMiddleware,
		taskEventMW.Middleware,
		middleware.SecurityHeadersMiddleware,
	).Wrap(r)
	if cfg.CSRFEnabled {
		handler = csrfmw.NewCSRFMiddleware(o.SessionSecret, cfg.HTTPHostname, version)(handler)
	}

	srv.Router = handler

	adminhandlers.ConfigFile = ConfigFile
	adminhandlers.Srv = srv
	adminhandlers.DBPool = dbPool
	adminhandlers.UpdateConfigKeyFunc = config.UpdateConfigKey

	emailProvider := o.EmailReg.ProviderFromConfig(cfg)
	if cfg.EmailEnabled && cfg.EmailProvider != "" && cfg.EmailFrom == "" {
		log.Printf("%s not set while EMAIL_PROVIDER=%s", config.EnvEmailFrom, cfg.EmailProvider)
	}

	dlqProvider := o.DLQReg.ProviderFromConfig(cfg, dbpkg.New(dbPool))

	workerCtx, workerCancel := context.WithCancel(context.Background())
	workers.Start(workerCtx, dbPool, emailProvider, dlqProvider, cfg, bus)
	srv.WorkerCancel = workerCancel

	return srv, nil
}
