package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

// Server bundles the application's configuration, router and runtime dependencies.
type Server struct {
	Config         config.RuntimeConfig
	Router         http.Handler
	Store          *sessions.CookieStore
	DB             *sql.DB
	SessionManager *core.SessionManager

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

// New returns a Server with the supplied dependencies.
func New(handler http.Handler, store *sessions.CookieStore, sm *core.SessionManager, db *sql.DB, cfg config.RuntimeConfig) *Server {
	return &Server{
		Config:         cfg,
		Router:         handler,
		Store:          store,
		SessionManager: sm,
		DB:             db,
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
