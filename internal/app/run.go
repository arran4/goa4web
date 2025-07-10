package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	imageshandler "github.com/arran4/goa4web/handlers/images"
	dbpkg "github.com/arran4/goa4web/internal/db"
	dbstart "github.com/arran4/goa4web/internal/dbstart"
	"github.com/arran4/goa4web/internal/dlq"
	email "github.com/arran4/goa4web/internal/email"
	emailutil "github.com/arran4/goa4web/internal/emailutil"
	"github.com/arran4/goa4web/internal/eventbus"
	middleware "github.com/arran4/goa4web/internal/middleware"
	csrfmw "github.com/arran4/goa4web/internal/middleware/csrf"
	notifications "github.com/arran4/goa4web/internal/notifications"
	routerpkg "github.com/arran4/goa4web/internal/router"
	startup "github.com/arran4/goa4web/internal/startup"
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
	store = sessions.NewCookieStore([]byte(sessionSecret))
	core.Store = store
	core.SessionName = sessionName
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	common.Version = version

	if err := startup.PerformChecks(cfg); err != nil {
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
	imageshandler.SetSigningKey(imageSignSecret)
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
		middleware.DBAdderMiddleware,
		middleware.CoreAdderMiddleware,
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
	if emailutil.EmailSendingEnabled() && cfg.EmailProvider != "" && cfg.EmailFrom == "" {
		log.Printf("%s not set while EMAIL_PROVIDER=%s", config.EnvEmailFrom, cfg.EmailProvider)
	}

	dlqProvider := dlq.ProviderFromConfig(cfg, dbpkg.New(dbPool))

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	startWorkers(workerCtx, dbPool, emailProvider, dlqProvider, cfg)

	if err := server.Run(ctx, srv, cfg.HTTPListen); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := eventbus.DefaultBus.Shutdown(shutdownCtx); err != nil {
		log.Printf("eventbus shutdown: %v", err)
	}
	workerCancel()

	return nil
}

// safeGo runs fn in a goroutine and terminates the program if a panic occurs.
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("goroutine panic: %v", r)
				os.Exit(1)
			}
		}()
		fn()
	}()
}

func startWorkers(ctx context.Context, db *sql.DB, provider email.Provider, dlqProvider dlq.DLQ, cfg config.RuntimeConfig) {
	log.Printf("Starting email worker")
	safeGo(func() {
		emailutil.EmailQueueWorker(ctx, dbpkg.New(db), provider, dlqProvider, time.Duration(cfg.EmailWorkerInterval)*time.Second)
	})
	log.Printf("Starting notification purger worker")
	safeGo(func() { notifications.NotificationPurgeWorker(ctx, dbpkg.New(db), time.Hour) })
	log.Printf("Starting event bus logger worker")
	safeGo(func() { eventbus.LogWorker(ctx, eventbus.DefaultBus) })
	log.Printf("Starting audit worker")
	safeGo(func() { eventbus.AuditWorker(ctx, eventbus.DefaultBus, dbpkg.New(db)) })
	log.Printf("Starting notification bus worker")
	safeGo(func() {
		notifications.BusWorker(ctx, eventbus.DefaultBus, notifications.Notifier{
			EmailProvider: provider,
			Queries:       dbpkg.New(db),
		}, dlqProvider)
	})
}
