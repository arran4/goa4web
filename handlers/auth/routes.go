package auth

import (
	"github.com/arran4/goa4web/handlers"
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the login and registration endpoints to r.
func RegisterRoutes(r *mux.Router) {
	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", registerTask.Page).Methods("GET").MatcherFunc(Not(handlers.RequiresAnAccount()))
	rr.HandleFunc("", registerTask.Action).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(registerTask.Matcher())

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", loginTask.Page).Methods("GET").MatcherFunc(Not(handlers.RequiresAnAccount()))
	lr.HandleFunc("", loginTask.Action).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(loginTask.Matcher())
	lr.HandleFunc("/verify", verifyPasswordTask.Action).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(verifyPasswordTask.Matcher())

	fr := r.PathPrefix("/forgot").Subrouter()
	fr.HandleFunc("", forgotPasswordTask.Page).Methods("GET").MatcherFunc(Not(handlers.RequiresAnAccount()))
	fr.HandleFunc("", forgotPasswordTask.Action).Methods("POST").MatcherFunc(Not(handlers.RequiresAnAccount())).MatcherFunc(forgotPasswordTask.Matcher())
}

// Register registers the auth router module.
func Register() {
	router.RegisterModule("auth", nil, RegisterRoutes)
}
