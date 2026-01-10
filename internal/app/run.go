package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/app/server"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	corelanguage "github.com/arran4/goa4web/core/language"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	feedsign "github.com/arran4/goa4web/internal/feedsign"
	imagesign "github.com/arran4/goa4web/internal/images"
	linksign "github.com/arran4/goa4web/internal/linksign"
	"github.com/arran4/goa4web/internal/middleware"
	csrfmw "github.com/arran4/goa4web/internal/middleware/csrf"
	nav "github.com/arran4/goa4web/internal/navigation"
	routerpkg "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// ConfigFile stores the path to the configuration file if provided on the
// command line. It is used by admin handlers when updating settings.
var ConfigFile string

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

// ServerOption configures additional parameters when constructing a server.
type ServerOption func(*serverOptions)

type serverOptions struct {
	SessionSecret   string
	ImageSignSecret string
	LinkSignSecret  string
	ShareSignSecret string
	APISecret       string
	DBReg           *dbdrivers.Registry
	EmailReg        *email.Registry
	DLQReg          *dlq.Registry
	TasksReg        *tasks.Registry
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

// WithLinkSignSecret supplies the external link signing secret.
func WithLinkSignSecret(secret string) ServerOption {
	return func(o *serverOptions) { o.LinkSignSecret = secret }
}

// WithShareSignSecret supplies the share signing secret.
func WithShareSignSecret(secret string) ServerOption {
	return func(o *serverOptions) { o.ShareSignSecret = secret }
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

// WithTasksRegistry sets the task registry.
func WithTasksRegistry(r *tasks.Registry) ServerOption {
	return func(o *serverOptions) { o.TasksReg = r }
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
func NewServer(ctx context.Context, cfg *config.RuntimeConfig, ah *adminhandlers.Handlers, opts ...ServerOption) (*server.Server, error) {
	o := &serverOptions{}
	for _, op := range opts {
		op(o)
	}

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
	sameSite := http.SameSiteStrictMode
	switch strings.ToLower(cfg.SessionSameSite) {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	}
	store.Options = &sessions.Options{Path: "/", HttpOnly: true, Secure: true, SameSite: sameSite}

	if o.DB == nil {
		var err error
		o.DB, err = PerformChecks(cfg, o.DBReg)
		if err != nil {
			return nil, fmt.Errorf("startup checks: %w", err)
		}
	}
	queries := db.New(o.DB)
	sm := db.NewSessionProxy(queries)
	if err := corelanguage.EnsureDefaultLanguage(context.Background(), queries, cfg.DefaultLanguage); err != nil {
		return nil, fmt.Errorf("ensure default language: %w", err)
	}
	if err := corelanguage.ValidateDefaultLanguage(context.Background(), queries, cfg.DefaultLanguage); err != nil {
		return nil, fmt.Errorf("default language: %w", err)
	}

	if err := config.ApplySMTPFallbacks(cfg); err != nil {
		return nil, fmt.Errorf("smtp fallback: %w", err)
	}

	imageSignExpiry, err := time.ParseDuration(cfg.ImageSignExpiry)
	if err != nil {
		return nil, fmt.Errorf("parsing image sign expiry: %w", err)
	}
	linkSignExpiry, err := time.ParseDuration(cfg.LinkSignExpiry)
	if err != nil {
		return nil, fmt.Errorf("parsing link sign expiry: %w", err)
	}
	shareSignExpiry, err := time.ParseDuration(cfg.ShareSignExpiry)
	if err != nil {
		return nil, fmt.Errorf("parsing share sign expiry: %w", err)
	}

	imgSigner := imagesign.NewSigner(cfg, o.ImageSignSecret, imageSignExpiry)
	linkSigner := linksign.NewSigner(cfg, o.LinkSignSecret, linkSignExpiry)
	shareSigner := sharesign.NewSigner(cfg, o.ShareSignSecret, shareSignExpiry)
	feedSigner := feedsign.NewSigner(cfg, o.LinkSignSecret)
	adminhandlers.AdminAPISecret = o.APISecret
	email.SetDefaultFromName(cfg.EmailFrom)

	if o.Bus == nil {
		o.Bus = eventbus.NewBus()
	}
	if o.RouterReg == nil {
		o.RouterReg = routerpkg.NewRegistry()
	}
	wsMod := websocket.NewModule(o.Bus, cfg)
	wsMod.Register(o.RouterReg)
	r := mux.NewRouter()

	navReg := nav.NewRegistry()
	routerpkg.RegisterRoutes(r, o.RouterReg, cfg, navReg)
	srv := server.New(
		server.WithStore(store),
		server.WithDB(o.DB),
		server.WithConfig(cfg),
		server.WithRouterRegistry(o.RouterReg),
		server.WithNavRegistry(navReg),
		server.WithDLQRegistry(o.DLQReg),
		server.WithTasksRegistry(o.TasksReg),
		server.WithBus(o.Bus),
		server.WithEmailRegistry(o.EmailReg),
		server.WithImageSigner(imgSigner),
		server.WithLinkSigner(linkSigner),
		server.WithShareSigner(shareSigner),
		server.WithFeedSigner(feedSigner),
		server.WithDBRegistry(o.DBReg),
		server.WithWebsocket(wsMod),
		server.WithTasksRegistry(o.TasksReg),
		server.WithSessionManager(sm),
	)
	share.RegisterShareRoutes(r, cfg, shareSigner)

	srv.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, r := srv.GetCoreData(w, r)
		if cd == nil {
			// GetCoreData handles all error reporting and redirects internally.
			// If it returns a nil CoreData, the response has already been sent
			// and we must return immediately to avoid a double write.
			return
		}
		if strings.HasPrefix(r.URL.Path, "/forum") {
			cd.NotFoundLink = &common.NotFoundLink{
				Text: "Go back to forum index",
				URL:  "/forum",
			}
		}
		handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
	})
	r.NotFoundHandler = srv.NotFoundHandler

	taskEventMW := middleware.NewTaskEventMiddleware(o.Bus)
	handler := middleware.NewMiddlewareChain(
		middleware.RecoverMiddleware,
		srv.CoreDataMiddleware(),
		middleware.RequestLoggerMiddleware,
		taskEventMW.Middleware,
		middleware.SecurityHeadersMiddleware,
	).Wrap(r)
	if cfg.CSRFEnabled {
		handler = csrfmw.NewCSRFMiddleware(o.SessionSecret, cfg.HTTPHostname, goa4web.Version)(handler)
	}

	srv.Router = handler

	if ah != nil {
		ah.ConfigFile = ConfigFile
		ah.Srv = srv
		ah.DBPool = o.DB
		ah.UpdateConfigKeyFunc = config.UpdateConfigKey
	}
	srv.TasksReg = o.TasksReg

	return srv, nil
}
