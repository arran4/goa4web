package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/workers"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	corelanguage "github.com/arran4/goa4web/core/language"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	email "github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	imagesign "github.com/arran4/goa4web/internal/images"
	middleware "github.com/arran4/goa4web/internal/middleware"
	csrfmw "github.com/arran4/goa4web/internal/middleware/csrf"
	routerpkg "github.com/arran4/goa4web/internal/router"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// ConfigFile stores the path to the configuration file if provided on the
// command line. It is used by admin handlers when updating settings.
var ConfigFile string

var (
	sessionName = "my-session"
	store       *sessions.CookieStore
	srv         *server.Server

	version = "dev"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

// RunWithConfig starts the application using the provided configuration and
// session secret. The context controls the lifetime of the HTTP server.
func RunWithConfig(ctx context.Context, cfg config.RuntimeConfig, sessionSecret, imageSignSecret string) error {
	log.Printf("application version %s starting", version)
	store = sessions.NewCookieStore([]byte(sessionSecret))
	core.Store = store
	core.SessionName = sessionName
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	if err := PerformChecks(cfg); err != nil {
		return fmt.Errorf("startup checks: %w", err)
	}

	dbPool := dbstart.GetDBPool()
	if err := corelanguage.ValidateDefaultLanguage(context.Background(), dbpkg.New(dbPool), cfg.DefaultLanguage); err != nil {
		return fmt.Errorf("default language: %w", err)
	}

	if err := config.ApplySMTPFallbacks(&cfg); err != nil {
		return fmt.Errorf("smtp fallback: %w", err)
	}
	config.AppRuntimeConfig = cfg
	imagesign.SetSigningKey(imageSignSecret)
	email.SetDefaultFromName(cfg.EmailFrom)

	if dbPool != nil {
		defer func() {
			if err := dbPool.Close(); err != nil {
				log.Printf("DB close error: %v", err)
			}
		}()
	}

	r := mux.NewRouter()
	routerpkg.RegisterRoutes(r)

	handler := middleware.NewMiddlewareChain(
		middleware.RecoverMiddleware,
		middleware.CoreAdderMiddlewareWithDB(dbPool, cfg.DBLogVerbosity),
		middleware.RequestLoggerMiddleware,
		middleware.TaskEventMiddleware,
		middleware.SecurityHeadersMiddleware,
	).Wrap(r)
	if csrfmw.CSRFEnabled() {
		handler = csrfmw.NewCSRFMiddleware(sessionSecret, cfg.HTTPHostname, version)(handler)
	}

	srv = server.New(handler, store, dbPool, cfg)
	adminhandlers.ConfigFile = ConfigFile
	adminhandlers.Srv = srv
	adminhandlers.DBPool = dbPool
	adminhandlers.UpdateConfigKeyFunc = config.UpdateConfigKey

	emailProvider := email.ProviderFromConfig(cfg)
	if config.EmailSendingEnabled() && cfg.EmailProvider != "" && cfg.EmailFrom == "" {
		log.Printf("%s not set while EMAIL_PROVIDER=%s", config.EnvEmailFrom, cfg.EmailProvider)
	}

	dlqProvider := dlq.ProviderFromConfig(cfg, dbpkg.New(dbPool))

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	workers.Start(workerCtx, dbPool, emailProvider, dlqProvider, cfg)

	if err := server.Run(ctx, srv, cfg.HTTPListen); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	log.Printf("application shutdown complete")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := eventbus.DefaultBus.Shutdown(shutdownCtx); err != nil {
		log.Printf("eventbus shutdown: %v", err)
	}
	workerCancel()

	return nil
}
