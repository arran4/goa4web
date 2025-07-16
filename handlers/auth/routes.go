package auth

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the login and registration endpoints to r.
func RegisterRoutes(r *mux.Router) {
	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", RegisterPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	rr.HandleFunc("", RegisterTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(RegisterTask.Match)

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", LoginUserPassPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	lr.HandleFunc("", LoginTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(LoginTask.Match)
	lr.HandleFunc("/verify", LoginVerifyPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(VerifyPasswordTask.Match)

	fr := r.PathPrefix("/forgot").Subrouter()
	fr.HandleFunc("", forgotPasswordTask.Page).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	fr.HandleFunc("", forgotPasswordTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(forgotPasswordTask.Match)
}

// Register registers the auth router module.
func Register() {
	router.RegisterModule("auth", nil, RegisterRoutes)
}
