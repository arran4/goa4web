package auth

import (
	"github.com/arran4/goa4web/handlers"
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the login and registration endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) []nav.RouterOptions {
	rr := r.PathPrefix("/register").Subrouter()
	rr.Use(handlers.IndexMiddleware(CustomIndex))
	rr.HandleFunc("", handlers.WithNoCache(registerTask.Page)).Methods("GET").MatcherFunc(Not(handlers.RequiresAnAccount()))
	rr.HandleFunc("", handlers.TaskHandler(registerTask)).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(registerTask.Matcher())

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", handlers.WithNoCache(loginTask.Page)).Methods("GET").MatcherFunc(Not(handlers.RequiresAnAccount()))
	lr.HandleFunc("", handlers.TaskHandler(loginTask)).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(loginTask.Matcher())
	lr.HandleFunc("/verify", handlers.TaskHandler(verifyPasswordTask)).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(verifyPasswordTask.Matcher())

	fr := r.PathPrefix("/forgot").Subrouter()
	fr.HandleFunc("", handlers.WithNoCache(forgotPasswordTask.Page)).Methods("GET").MatcherFunc(Not(handlers.RequiresAnAccount()))
	fr.HandleFunc("", handlers.TaskHandler(emailAssociationRequestTask)).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(emailAssociationRequestTask.Matcher())
	fr.HandleFunc("", handlers.TaskHandler(forgotPasswordTask)).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(forgotPasswordTask.Matcher())
	return nil
}

// Register registers the auth router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("auth", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []nav.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
