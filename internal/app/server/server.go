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
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	router "github.com/arran4/goa4web/internal/router"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/middleware"
	nav "github.com/arran4/goa4web/internal/navigation"
)

// Server bundles the application's configuration, router and runtime dependencies.
type Server struct {
	RouterReg *router.Registry
	Nav    *navigation.Registry
	Config      config.RuntimeConfig
	Router      http.Handler
	Store       *sessions.CookieStore
	DB          *sql.DB
	Bus         *eventbus.Bus
	EmailReg    *email.Registry
	ImageSigner *imagesign.Signer

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

// New returns a Server with the supplied dependencies.
func New(handler http.Handler, store *sessions.CookieStore, db *sql.DB, cfg config.RuntimeConfig, reg *router.Registry, nav *navigation.Registry) *Server {
	return &Server{
		Config:    cfg,
		Router:    handler,
		Store:     store,
		DB:        db,
		RouterReg: reg,
		Nav:    nav,
  }
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
			cd := common.NewCoreData(r.Context(), queries,
				common.WithImageSigner(s.ImageSigner),
				common.WithSession(session),
				common.WithEmailProvider(provider),
				common.WithAbsoluteURLBase(base),
				common.WithConfig(s.Config),
				common.WithSessionManager(sm))
			cd.UserID = uid
			_ = cd.UserRoles()

			idx := nav.IndexItems()
			cd.IndexItems = idx
			cd.Title = "Arran's Site"
			cd.FeedsEnabled = s.Config.FeedsEnabled
			cd.AdminMode = r.URL.Query().Get("mode") == "admin"
			if uid != 0 && s.Config.NotificationsEnabled {
				cd.NotificationCount = int32(cd.UnreadNotificationCount())
			}
			ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
	return Run(ctx, s, s.Config.HTTPListen)
}
