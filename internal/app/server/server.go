package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	imagesign "github.com/arran4/goa4web/internal/images"
	linksign "github.com/arran4/goa4web/internal/linksign"
	"github.com/arran4/goa4web/internal/middleware"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/websocket"
	"github.com/arran4/goa4web/workers"
)

// Server bundles the application's configuration, router and runtime dependencies.
type Server struct {
	RouterReg   *router.Registry
	Nav         *nav.Registry
	Config      *config.RuntimeConfig
	Router      http.Handler
	Store       *sessions.CookieStore
	DB          *sql.DB
	Bus         *eventbus.Bus
	EmailReg    *email.Registry
	ImageSigner *imagesign.Signer
	LinkSigner  *linksign.Signer
	TasksReg    *tasks.Registry
	DBReg       *dbdrivers.Registry
	DLQReg      *dlq.Registry
	Websocket   *websocket.Module

	WorkerCancel context.CancelFunc

	addr       string
	httpServer *http.Server
}

// Addr returns the address the server is listening on after Start is called.
func (s *Server) Addr() string { return s.addr }

// Start begins serving HTTP requests on the given address. If the port is
// specified as :0, the automatically chosen address can be retrieved via Addr().
func (s *Server) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	s.addr = ln.Addr().String()
	s.httpServer = &http.Server{Handler: s.Router}
	log.Printf("Server started on http://%s", s.addr)
	if err := s.httpServer.Serve(ln); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

// Close releases resources associated with the server such as background
// workers, the event bus and the database connection.
func (s *Server) Close() {
	if s.WorkerCancel != nil {
		s.WorkerCancel()
	}
	if s.Bus != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Bus.Shutdown(shutdownCtx); err != nil {
			log.Printf("eventbus shutdown: %v", err)
		}
	}
	if s.DB != nil {
		if err := s.DB.Close(); err != nil {
			log.Printf("DB close error: %v", err)
		}
	}
}

// Option configures the Server returned by New.
type Option func(*Server)

// WithHandler sets the HTTP handler used by the server.
func WithHandler(h http.Handler) Option { return func(s *Server) { s.Router = h } }

// WithStore sets the session store used by the server.
func WithStore(store *sessions.CookieStore) Option { return func(s *Server) { s.Store = store } }

// WithDB sets the database pool.
func WithDB(db *sql.DB) Option { return func(s *Server) { s.DB = db } }

// WithConfig supplies the runtime configuration.
func WithConfig(cfg *config.RuntimeConfig) Option { return func(s *Server) { s.Config = cfg } }

// WithRouterRegistry sets the router registry.
func WithRouterRegistry(r *router.Registry) Option { return func(s *Server) { s.RouterReg = r } }

// WithNavRegistry sets the navigation registry.
func WithNavRegistry(n *nav.Registry) Option { return func(s *Server) { s.Nav = n } }

// WithDLQRegistry sets the dead letter queue registry.
func WithDLQRegistry(r *dlq.Registry) Option { return func(s *Server) { s.DLQReg = r } }

// WithBus sets the event bus used by the server.
func WithBus(b *eventbus.Bus) Option { return func(s *Server) { s.Bus = b } }

// WithEmailRegistry sets the email provider registry.
func WithEmailRegistry(r *email.Registry) Option { return func(s *Server) { s.EmailReg = r } }

// WithImageSigner sets the image signer.
func WithImageSigner(signer *imagesign.Signer) Option {
	return func(s *Server) { s.ImageSigner = signer }
}

// WithLinkSigner sets the external link signer.
func WithLinkSigner(signer *linksign.Signer) Option {
	return func(s *Server) { s.LinkSigner = signer }
}

// WithDBRegistry sets the database driver registry.
func WithDBRegistry(r *dbdrivers.Registry) Option { return func(s *Server) { s.DBReg = r } }

// WithWebsocket sets the websocket module.
func WithWebsocket(w *websocket.Module) Option { return func(s *Server) { s.Websocket = w } }

// WithTasksRegistry sets the tasks registry used by the server.
func WithTasksRegistry(r *tasks.Registry) Option { return func(s *Server) { s.TasksReg = r } }

// New returns a Server configured using the supplied options.
func New(opts ...Option) *Server {
	s := &Server{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// CoreDataMiddleware constructs the middleware responsible for populating
// CoreData in the request context using the server's configured dependencies.
func (s *Server) CoreDataMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := core.GetSession(r)
			if err != nil {
				core.SessionErrorRedirect(w, r, err)
				return
			}
			var uid int32
			if v, ok := session.Values["UID"].(int32); ok {
				uid = v
			}
			if expi, ok := session.Values["ExpiryTime"]; ok {
				var exp int64
				switch t := expi.(type) {
				case int64:
					exp = t
				case int:
					exp = int64(t)
				case float64:
					exp = int64(t)
				}
				if exp != 0 && time.Now().Unix() > exp {
					delete(session.Values, "UID")
					delete(session.Values, "LoginTime")
					delete(session.Values, "ExpiryTime")
					middleware.RedirectToLogin(w, r, session)
					return
				}
			}
			if s.DB == nil {
				ue := common.UserError{Err: fmt.Errorf("db not initialized"), ErrorMessage: "database unavailable"}
				log.Printf("%s: %v", ue.ErrorMessage, ue.Err)
				http.Error(w, ue.ErrorMessage, http.StatusInternalServerError)
				return
			}

			queries := dbpkg.New(s.DB)
			sm := queries
			if s.Config.DBLogVerbosity > 0 {
				log.Printf("db pool stats: %+v", s.DB.Stats())
			}

			if session.ID != "" {
				if uid != 0 {
					if err := queries.InsertSession(r.Context(), dbpkg.InsertSessionParams{SessionID: session.ID, UsersIdusers: uid}); err != nil {
						log.Printf("insert session: %v", err)
					}
				} else {
					if err := queries.DeleteSessionByID(r.Context(), session.ID); err != nil {
						log.Printf("delete session: %v", err)
					}
				}
			}

			base := "http://" + r.Host
			if s.Config.HTTPHostname != "" {
				base = strings.TrimRight(s.Config.HTTPHostname, "/")
			}
			provider := s.EmailReg.ProviderFromConfig(s.Config)
			cd := common.NewCoreData(r.Context(), queries, s.Config,
				common.WithImageSigner(s.ImageSigner),
				common.WithLinkSigner(s.LinkSigner),
				common.WithSession(session),
				common.WithEmailProvider(provider),
				common.WithAbsoluteURLBase(base),
				common.WithSessionManager(sm),
				common.WithNavRegistry(s.Nav),
				common.WithTasksRegistry(s.TasksReg),
				common.WithDBRegistry(s.DBReg),
			)
			cd.UserID = uid
			_ = cd.UserRoles()

			if s.Nav != nil {
				cd.IndexItems = s.Nav.IndexItems()
			}
			cd.Title = "Arran's Site"
			cd.FeedsEnabled = s.Config.FeedsEnabled
			cd.AdminMode = r.URL.Query().Get("mode") == "admin"
			if strings.HasPrefix(r.URL.Path, "/admin") && cd.HasRole("administrator") {
				cd.AdminMode = true
			}
			if uid != 0 && s.Config.NotificationsEnabled {
				cd.NotificationCount = int32(cd.UnreadNotificationCount())
			}
			ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// startWorkers launches the background workers when the server starts.
func (s *Server) startWorkers(ctx context.Context) {
	if s.WorkerCancel != nil {
		return
	}
	workerCtx, cancel := context.WithCancel(ctx)
	emailProvider := s.EmailReg.ProviderFromConfig(s.Config)
	if s.Config.EmailEnabled && s.Config.EmailProvider != "" && s.Config.EmailFrom == "" {
		log.Printf("%s not set while EMAIL_PROVIDER=%s", config.EnvEmailFrom, s.Config.EmailProvider)
	}
	dlqProvider := s.DLQReg.ProviderFromConfig(s.Config, dbpkg.New(s.DB))
	workers.Start(workerCtx, s.DB, emailProvider, dlqProvider, s.Config, s.Bus)
	s.WorkerCancel = cancel
}

// Run starts the HTTP server and blocks until the context is cancelled.
func Run(ctx context.Context, srv *Server, addr string) error {
	go func() {
		if err := srv.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()
	<-ctx.Done()
	log.Printf("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}
	return nil
}

// RunContext starts the server using the configured HTTPListen address.
func (s *Server) RunContext(ctx context.Context) error {
	s.startWorkers(ctx)
	return Run(ctx, s, s.Config.HTTPListen)
}
