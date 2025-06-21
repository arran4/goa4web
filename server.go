package main

import (
	"context"
	"database/sql"
	"github.com/gorilla/sessions"
	"log"
	"net"
	"net/http"
)

// Server bundles the application's configuration, router and runtime dependencies.
type Server struct {
	DBConfig    DBConfig
	EmailConfig EmailConfig
	Router      http.Handler
	Store       *sessions.CookieStore
	DB          *sql.DB

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
		return err
	}
	s.addr = ln.Addr().String()
	s.httpServer = &http.Server{Handler: s.Router}
	log.Printf("Server started on http://%s", s.addr)
	return s.httpServer.Serve(ln)
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
