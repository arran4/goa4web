package auth

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	router "github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the login and registration endpoints to r.
func RegisterRoutes(r *mux.Router) {
	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", registerTask.Page).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	rr.HandleFunc("", registerTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(registerTask.Match)

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", loginTask.Page).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	lr.HandleFunc("", loginTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(loginTask.Match)
	lr.HandleFunc("/verify", verifyPasswordTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(verifyPasswordTask.Match)

	fr := r.PathPrefix("/forgot").Subrouter()
	fr.HandleFunc("", forgotPasswordTask.Page).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	fr.HandleFunc("", forgotPasswordTask.Action).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(forgotPasswordTask.Match)
}

// Register registers the auth router module.
func Register() {
	router.RegisterModule("auth", nil, RegisterRoutes)
}
