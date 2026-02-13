package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/middleware"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/stats"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/websocket"
	"github.com/arran4/goa4web/workers"
)

// Server bundles the application's configuration, router and runtime dependencies.
type Server struct {
	RouterReg       *router.Registry
	Nav             *nav.Registry
	Config          *config.RuntimeConfig
	ConfigFile      string
	Router          http.Handler
	NotFoundHandler http.Handler
	Store           *sessions.CookieStore
	DB              *sql.DB
	Queries         db.Querier
	Bus             *eventbus.Bus
	EmailReg        *email.Registry

	// Signing keys (simple strings, no complex objects)
	FeedSignKey  string
	ImageSignKey string
	LinkSignKey  string
	ShareSignKey string

	SessionManager common.SessionManager
	TasksReg       *tasks.Registry
	DBReg          *dbdrivers.Registry
	DLQReg         *dlq.Registry
	Websocket      *websocket.Module

	WorkerCancel context.CancelFunc

	addr       string
	httpServer *http.Server

	cachedEmailProvider common.MailProvider
	cachedEmailError    error
	lastEmailConfig     *config.RuntimeConfig
	emailMu             sync.Mutex
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

// WithQuerier sets a custom query implementation for the server.
func WithQuerier(q db.Querier) Option { return func(s *Server) { s.Queries = q } }

// WithConfig supplies the runtime configuration.
func WithConfig(cfg *config.RuntimeConfig) Option { return func(s *Server) { s.Config = cfg } }

// WithConfigFile sets the config file path.
func WithConfigFile(path string) Option { return func(s *Server) { s.ConfigFile = path } }

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

// WithImageSignKey sets the image signing key.
func WithImageSignKey(key string) Option {
	return func(s *Server) { s.ImageSignKey = key }
}

// WithLinkSignKey sets the external link signing key.
func WithLinkSignKey(key string) Option {
	return func(s *Server) { s.LinkSignKey = key }
}

// WithShareSignKey sets the share signing key.
func WithShareSignKey(key string) Option {
	return func(s *Server) { s.ShareSignKey = key }
}

// WithFeedSignKey sets the feed signing key.
func WithFeedSignKey(key string) Option {
	return func(s *Server) { s.FeedSignKey = key }
}

// WithSessionManager sets the session manager used by the server.
func WithSessionManager(sm common.SessionManager) Option {
	return func(s *Server) { s.SessionManager = sm }
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

func (s *Server) getEmailProvider() (common.MailProvider, error) {
	s.emailMu.Lock()
	defer s.emailMu.Unlock()
	if s.lastEmailConfig == s.Config {
		return s.cachedEmailProvider, s.cachedEmailError
	}
	p, err := s.EmailReg.ProviderFromConfig(s.Config)
	s.cachedEmailProvider = p
	s.cachedEmailError = err
	s.lastEmailConfig = s.Config
	return p, err
}

// CoreDataMiddleware constructs the middleware responsible for populating
// CoreData in the request context using the server's configured dependencies.
func (s *Server) GetCoreData(w http.ResponseWriter, r *http.Request) (*common.CoreData, *http.Request) {
	session, err := core.GetSession(r)
	if err != nil {
		core.SessionErrorRedirect(w, r, err)
		return nil, nil
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
			_ = middleware.RedirectToLogin(w, r, session)
			return nil, nil
		}
	}
	queries := s.Queries
	if queries == nil {
		if s.DB == nil {
			ue := common.UserError{Err: fmt.Errorf("db not initialized"), ErrorMessage: "database unavailable"}
			log.Printf("%s: %v", ue.ErrorMessage, ue.Err)
			handlers.RenderErrorPage(w, r, errors.New(ue.ErrorMessage))
			return nil, nil
		}
		queries = db.New(s.DB)
	}

	sm := s.SessionManager
	if sm == nil {
		sm = db.NewSessionProxy(queries)
	}
	if s.Config.DBLogVerbosity > 0 && s.DB != nil {
		log.Printf("db pool stats: %+v", s.DB.Stats())
	}

	if session.ID != "" && sm != nil {
		if uid != 0 {
			if err := sm.InsertSession(r.Context(), session.ID, uid); err != nil {
				log.Printf("insert session: %v", err)
			}
		} else {
			if err := sm.DeleteSessionByID(r.Context(), session.ID); err != nil {
				log.Printf("delete session: %v", err)
			}
		}
	}

	base := "http://" + r.Host
	if s.Config.BaseURL != "" {
		base = strings.TrimRight(s.Config.BaseURL, "/")
	}
	provider, providerErr := s.getEmailProvider()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	modules := []string{}
	if s.RouterReg != nil {
		modules = s.RouterReg.Names()
	}
	customQueries, _ := queries.(db.CustomQueries)
	cd := common.NewCoreData(r.Context(), queries, s.Config,
		common.WithImageSignKey(s.ImageSignKey),
		common.WithCustomQueries(customQueries),
		common.WithLinkSignKey(s.LinkSignKey),
		common.WithShareSignKey(s.ShareSignKey),
		common.WithFeedSignKey(s.FeedSignKey),
		common.WithEventBus(s.Bus),
		common.WithSession(session),
		common.WithEmailProvider(provider),
		common.WithAbsoluteURLBase(base),
		common.WithSessionManager(sm),
		common.WithNavRegistry(s.Nav),
		common.WithTasksRegistry(s.TasksReg),
		common.WithDLQRegistry(s.DLQReg),
		common.WithDBRegistry(s.DBReg),
		common.WithEmailRegistry(s.EmailReg),
		common.WithRouterModules(modules),
		common.WithOffset(offset),
		common.WithSiteTitle("Arran's Site"),
	)
	if providerErr != nil {
		cd.EmailProviderError = providerErr.Error()
	}
	cd.UserID = uid
	_ = cd.UserRoles()

	cd.AdminMode = r.URL.Query().Get("mode") == "admin"
	if strings.HasPrefix(r.URL.Path, "/admin") && cd.HasAdminRole() {
		cd.AdminMode = true
	}
	if s.Nav != nil {
		cd.IndexItems = s.Nav.IndexItemsWithPermission(func(section, item string) bool {
			return cd.HasGrant(section, item, "view", 0) || cd.IsAdmin()
		})
	}
	cd.FeedsEnabled = s.Config.FeedsEnabled
	if uid != 0 && s.Config.NotificationsEnabled {
		cd.NotificationCount = int32(cd.UnreadNotificationCount())
	}
	ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
	return cd, r.WithContext(ctx)
}

func (s *Server) CoreDataMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, rCtx := s.GetCoreData(w, r); rCtx != nil {
				next.ServeHTTP(w, rCtx)
			}
		})
	}
}

// startWorkers launches the background workers when the server starts.
func (s *Server) startWorkers(ctx context.Context) {
	if s.WorkerCancel != nil {
		return
	}
	emailProvider, err := s.EmailReg.ProviderFromConfig(s.Config)
	if err != nil {
		log.Printf("Email provider init failed: %v", err)
	}
	if s.Config.EmailEnabled && s.Config.EmailProvider != "" && s.Config.EmailFrom == "" {
		log.Printf("%s not set while EMAIL_PROVIDER=%s", config.EnvEmailFrom, s.Config.EmailProvider)
	}
	var q db.Querier
	if s.Queries != nil {
		q = s.Queries
	} else if s.DB != nil {
		q = db.New(s.DB)
	} else {
		log.Printf("startWorkers: no db or querier available")
		return
	}
	workerCtx, cancel := context.WithCancel(ctx)
	dlqProvider := s.DLQReg.ProviderFromConfig(s.Config, q)
	workers.Start(workerCtx, q, emailProvider, dlqProvider, s.Config, s.Bus)
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

	modules := []string{}
	if srv.RouterReg != nil {
		modules = srv.RouterReg.Names()
	}
	data := stats.BuildServerStatsData(srv.Config, srv.ConfigFile, srv.TasksReg, srv.DBReg, srv.DLQReg, srv.EmailReg, modules)
	if b, err := json.Marshal(data); err == nil {
		log.Printf("Server stats: %s", string(b))
	} else {
		log.Printf("Server stats error: %v", err)
	}

	if srv.DB != nil {
		queries := db.New(srv.DB)
		var q db.Querier = queries
		customQueries, _ := q.(db.CustomQueries)
		usageTimeout := 5 * time.Minute
		ctx, cancel := context.WithTimeout(context.Background(), usageTimeout)
		defer cancel()
		usageData := stats.BuildUsageStatsData(ctx, queries, customQueries, srv.Config.StatsStartYear)
		if b, err := json.Marshal(usageData); err == nil {
			log.Printf("Usage stats: %s", string(b))
		} else {
			log.Printf("Usage stats error: %v", err)
		}
	}

	return nil
}

// RunContext starts the server using the configured HTTPListen address.
func (s *Server) RunContext(ctx context.Context) error {
	s.startWorkers(ctx)
	return Run(ctx, s, s.Config.HTTPListen)
}
