package auth

import (
	"github.com/arran4/goa4web/handlers"
	gml "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the login and registration endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) []nav.RouterOptions {
	rr := r.PathPrefix("/register").Subrouter()
	rr.Use(handlers.IndexMiddleware(CustomIndex))
	rr.HandleFunc("", handlers.WithNoCache(registerTask.Page)).Methods("GET").MatcherFunc(gml.Not(handlers.RequiresAnAccount()))
	rr.HandleFunc("", handlers.TaskHandler(registerTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(registerTask.Matcher())

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", handlers.WithNoCache(loginTask.Page)).Methods("GET").MatcherFunc(gml.Not(handlers.RequiresAnAccount()))
	lr.HandleFunc("", handlers.TaskHandler(loginTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(loginTask.Matcher())
	lr.HandleFunc("/verify", handlers.TaskHandler(verifyPasswordTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(verifyPasswordTask.Matcher())

	lr.HandleFunc("/passkey/begin", handlers.WithNoCache(loginPasskeyBegin)).Methods("GET")
	lr.HandleFunc("/passkey/finish", loginPasskeyFinish).Methods("POST")

	fr := r.PathPrefix("/forgot").Subrouter()
	fr.HandleFunc("", handlers.WithNoCache(forgotPasswordTask.Page)).Methods("GET").MatcherFunc(gml.Not(handlers.RequiresAnAccount()))
	fr.HandleFunc("", handlers.TaskHandler(emailAssociationRequestTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(emailAssociationRequestTask.Matcher())
	fr.HandleFunc("", handlers.TaskHandler(forgotPasswordTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(forgotPasswordTask.Matcher())
	return nil
}

// Register registers the auth router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("auth", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []nav.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
